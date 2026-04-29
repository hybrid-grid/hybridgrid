package scheduler

import (
	"crypto/rand"
	"encoding/binary"
	"math"
	"sync"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// EpsilonGreedyScheduler is the canonical bandit baseline from
// Sutton & Barto 2018, "Reinforcement Learning: An Introduction" §2.2-2.4.
//
// Per-worker action-value Q(a) is maintained as the sample mean of
// observed rewards, updated incrementally per §2.4:
//
//	Q_{n+1} = Q_n + (R_n - Q_n) / n
//
// Selection is ε-greedy: with probability ε the scheduler picks a
// uniform random eligible worker (explore); otherwise it picks the
// eligible worker with the highest Q (exploit). Cold workers (n == 0)
// are treated as Q = 0; ε-greedy exploration eventually probes them.
//
// This is intentionally feature-blind. It serves as the explicit
// bandit baseline against which the LinUCB scheduler (M3) demonstrates
// the value of feature-conditioned policies, per Li et al. 2010 §5.
type EpsilonGreedyScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker
	epsilon        float64

	mu     sync.Mutex
	values map[string]*armState
}

type armState struct {
	q float64 // running mean reward
	n int64   // number of samples
}

// EpsilonGreedyConfig holds construction parameters. Epsilon must be in
// [0, 1]; values outside that range are clamped to the nearest endpoint
// and a warning is *not* emitted because the caller (CLI or factory) is
// expected to validate before reaching this point.
type EpsilonGreedyConfig struct {
	Registry       registry.Registry
	CircuitChecker CircuitChecker
	// Epsilon is the exploration rate. Default 0.1 follows Sutton & Barto
	// fig. 2.2 baseline. Must be in [0, 1].
	Epsilon float64
}

// NewEpsilonGreedyScheduler constructs the scheduler. If Epsilon is
// zero the scheduler is purely greedy (no exploration); if it is one
// every selection is uniform random.
func NewEpsilonGreedyScheduler(cfg EpsilonGreedyConfig) *EpsilonGreedyScheduler {
	eps := cfg.Epsilon
	if eps < 0 {
		eps = 0
	}
	if eps > 1 {
		eps = 1
	}
	return &EpsilonGreedyScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
		epsilon:        eps,
		values:         make(map[string]*armState),
	}
}

// Select implements the base Scheduler interface; it delegates through
// SelectWithDispatchInfo and discards the introspection payload.
func (s *EpsilonGreedyScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS)
	return w, err
}

// SelectWithDispatchInfo implements LearningScheduler. The returned
// DispatchInfo's QValueAtDispatch is the chosen worker's current
// running-mean reward estimate (zero for cold workers), and
// WasExploration reports whether the choice was random.
func (s *EpsilonGreedyScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := s.eligibleWorkers(buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}
	if len(candidates) == 1 {
		// Fast path: no choice, no exploration cost. Avoids the
		// 1-worker overhead the M1 P2C benchmark exhibited.
		w := candidates[0]
		return w, DispatchInfo{QValueAtDispatch: s.qValue(w.ID), WasExploration: false}, nil
	}

	exploring := s.epsilon > 0 && uniformFloat() < s.epsilon
	var chosen *registry.WorkerInfo
	if exploring {
		chosen = candidates[uniformInt(len(candidates))]
	} else {
		chosen = s.argmaxQ(candidates)
	}
	return chosen, DispatchInfo{QValueAtDispatch: s.qValue(chosen.ID), WasExploration: exploring}, nil
}

// RecordOutcome implements LearningScheduler. It updates Q(a) using the
// incremental sample-mean formula. Failed tasks still update the
// estimator: a worker that consistently fails should have its Q drop
// (assuming the caller passes a punishing reward on failure).
func (s *EpsilonGreedyScheduler) RecordOutcome(workerID string, reward float64, _ bool) {
	if workerID == "" || math.IsNaN(reward) || math.IsInf(reward, 0) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.values[workerID]
	if !ok {
		st = &armState{}
		s.values[workerID] = st
	}
	st.n++
	// Sutton & Barto §2.4: Q_{n+1} = Q_n + (R_n - Q_n)/n
	st.q += (reward - st.q) / float64(st.n)
}

// qValue returns the current Q estimate for a worker, or 0 if no
// samples have been recorded.
func (s *EpsilonGreedyScheduler) qValue(workerID string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.values[workerID]; ok {
		return st.q
	}
	return 0
}

// argmaxQ returns the candidate with the highest Q. Ties are broken by
// taking the first encountered (input order), which is the registry's
// listing order — deterministic for a given cluster snapshot.
func (s *EpsilonGreedyScheduler) argmaxQ(candidates []*registry.WorkerInfo) *registry.WorkerInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	var best *registry.WorkerInfo
	bestQ := math.Inf(-1)
	for _, w := range candidates {
		q := 0.0
		if st, ok := s.values[w.ID]; ok {
			q = st.q
		}
		if q > bestQ {
			best = w
			bestQ = q
		}
	}
	return best
}

// eligibleWorkers applies the same admission rules as P2CScheduler so
// the comparison in M3 evaluation is apples-to-apples.
func (s *EpsilonGreedyScheduler) eligibleWorkers(buildType pb.BuildType, arch pb.Architecture, clientOS string) ([]*registry.WorkerInfo, error) {
	workers := s.registry.ListByCapability(buildType, arch)
	if len(workers) == 0 {
		if s.registry.Count() == 0 {
			return nil, ErrNoWorkers
		}
		return nil, ErrNoMatchingWorkers
	}
	if clientOS != "" {
		workers = filterByOS(workers, clientOS)
		if len(workers) == 0 {
			return nil, ErrNoMatchingWorkers
		}
	}

	candidates := make([]*registry.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		if w.State == registry.WorkerStateUnhealthy {
			continue
		}
		if s.circuitChecker != nil && s.circuitChecker.IsOpen(w.ID) {
			continue
		}
		maxParallel := w.MaxParallel
		if maxParallel <= 0 {
			maxParallel = 4
		}
		if w.ActiveTasks >= maxParallel {
			continue
		}
		candidates = append(candidates, w)
	}
	if len(candidates) == 0 {
		// Relax the circuit-breaker filter so a strictly-busy cluster
		// still returns *something*; mirrors P2CScheduler's behaviour.
		for _, w := range workers {
			if w.State == registry.WorkerStateUnhealthy {
				continue
			}
			maxParallel := w.MaxParallel
			if maxParallel <= 0 {
				maxParallel = 4
			}
			if w.ActiveTasks < maxParallel {
				candidates = append(candidates, w)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, ErrNoMatchingWorkers
	}
	return candidates, nil
}

// uniformFloat returns a draw from U(0, 1) using crypto/rand. Slow but
// reuses the existing project's randomness source (P2CScheduler's
// pickTwo also uses crypto/rand). The cost is negligible compared to a
// gRPC dispatch.
func uniformFloat() float64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0
	}
	// Map to [0, 1). Top 53 bits → IEEE-754 mantissa precision.
	u := binary.BigEndian.Uint64(b[:]) >> 11
	return float64(u) / (1 << 53)
}

// uniformInt returns a uniform integer in [0, n). Falls back to 0 on
// rand failure (same convention as the existing cryptoRandInt helper).
func uniformInt(n int) int {
	if n <= 0 {
		return 0
	}
	return cryptoRandInt(n)
}
