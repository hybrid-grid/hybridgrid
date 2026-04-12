# v0.3.0 Observability - Unresolved Problems

**Session**: Atlas (Orchestrator)  
**Started**: 2026-03-17  
**Purpose**: Track blockers requiring escalation or architectural decisions

---

## Active Blockers

[Blockers will be documented here if encountered]

---

## Potential Risks

### High Risk Items
- **Task 1.1.4** (worker_latency_ms): Need to find where worker RPCs happen in scheduler
- **Task 1.1.5** (circuit_state): Need to find circuit breaker integration points
- **Task 1.2** (OTel/TLS wiring): Potential config conflicts or missing dependencies

### Medium Risk Items
- **Task 1.3** (worker metrics): HTTP server might conflict with existing worker setup
- **Task 2.4** (dashboard API): Need to locate dashboard handler code

### Low Risk Items
- **Task 2.1-2.3**: Straightforward implementations with clear solutions

---

## Escalation Log

[Problems requiring oracle/metis consultation will be documented here]
