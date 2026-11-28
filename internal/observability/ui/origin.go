package ui

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// OriginWildcard as the sole-purpose allowlist entry disables the
// WebSocket origin check entirely. It exists so operators can opt back
// in to the old allow-all behaviour explicitly.
const OriginWildcard = "*"

// originRejectSampler keeps a misbehaving page that reconnects in a
// tight loop from flooding the log with rejection warnings.
var originRejectSampler = &zerolog.BurstSampler{Burst: 5, Period: time.Minute}

// newOriginChecker returns a websocket.Upgrader CheckOrigin function.
//
// It follows gorilla/websocket's default policy: a request without an
// Origin header (CLI tools, Go clients) is accepted, and a request whose
// Origin host:port equals the request Host (compared case-insensitively)
// is accepted. On top of that, any Origin listed in allowed is accepted,
// and a single "*" entry accepts everything. Malformed Origin values are
// rejected.
func newOriginChecker(allowed []string) func(r *http.Request) bool {
	allowAll := false
	allowSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		switch o {
		case "":
			continue
		case OriginWildcard:
			allowAll = true
		default:
			allowSet[normalizeOrigin(o)] = struct{}{}
		}
	}

	return func(r *http.Request) bool {
		if allowAll {
			return true
		}
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		if _, ok := allowSet[normalizeOrigin(origin)]; ok {
			return true
		}
		if u, err := url.Parse(origin); err == nil && u.Host != "" && strings.EqualFold(u.Host, r.Host) {
			return true
		}
		sampled := log.Sample(originRejectSampler)
		sampled.Warn().
			Str("origin", origin).
			Str("host", r.Host).
			Str("remote", r.RemoteAddr).
			Msg("rejected WebSocket upgrade from disallowed origin; use --allowed-origins to permit it")
		return false
	}
}

// normalizeOrigin lowercases an origin and drops a trailing slash so
// "HTTP://Localhost:5173/" and "http://localhost:5173" compare equal.
func normalizeOrigin(o string) string {
	return strings.TrimRight(strings.ToLower(strings.TrimSpace(o)), "/")
}
