package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/h3nr1-d14z/hybridgrid/internal/security/auth"
)

// LogLevelRequest represents a log level change request.
type LogLevelRequest struct {
	Level string `json:"level"`
}

// LogLevelResponse represents the response for log level operations.
type LogLevelResponse struct {
	Level    string `json:"level"`
	Previous string `json:"previous,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// validLevels is the set of valid zerolog levels.
var validLevels = map[string]zerolog.Level{
	"trace": zerolog.TraceLevel,
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
	"fatal": zerolog.FatalLevel,
	"panic": zerolog.PanicLevel,
}

// NewLogLevelHandler creates an HTTP handler for managing log levels.
// Supports:
//   - GET /log-level: returns current log level as JSON
//   - PUT /log-level or POST /log-level: changes log level from JSON body
//
// When token is non-empty, every request must present it in the
// Authorization header ("Authorization: Bearer <token>"; the Bearer
// prefix is optional). An empty token keeps the endpoint open — the
// unauthenticated LAN default. Validation is constant-time via
// auth.ValidateToken, the same primitive the gRPC interceptors use.
func NewLogLevelHandler(token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if token != "" && !authorized(r, token) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "unauthorized"}) //nolint:errcheck
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetLogLevel(w, r)
		case http.MethodPut, http.MethodPost:
			handleSetLogLevel(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"}) //nolint:errcheck
		}
	})
}

// handleGetLogLevel handles GET /log-level requests.
func handleGetLogLevel(w http.ResponseWriter, r *http.Request) {
	currentLevel := zerolog.GlobalLevel()
	levelName := currentLevel.String()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LogLevelResponse{Level: levelName}) //nolint:errcheck
}

// handleSetLogLevel handles PUT/POST /log-level requests.
func handleSetLogLevel(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"}) //nolint:errcheck
		return
	}
	defer r.Body.Close()

	var req LogLevelRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"}) //nolint:errcheck
		return
	}

	// Validate level
	newLevel, ok := validLevels[req.Level]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("invalid level: %q (must be one of: trace, debug, info, warn, error, fatal, panic)", req.Level),
		}) //nolint:errcheck
		return
	}

	// Get previous level
	previousLevel := zerolog.GlobalLevel()
	previousLevelName := previousLevel.String()

	// Set new level
	zerolog.SetGlobalLevel(newLevel)

	// Log the change
	log.Info().
		Str("from", previousLevelName).
		Str("to", req.Level).
		Msg("Log level changed")

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LogLevelResponse{
		Level:    req.Level,
		Previous: previousLevelName,
	}) //nolint:errcheck
}

// authorized checks the request's Authorization header against the
// expected token. The RFC 7235 Bearer scheme prefix is accepted and
// optional, so both `Authorization: $TOKEN` and
// `Authorization: Bearer $TOKEN` work.
func authorized(r *http.Request, token string) bool {
	provided := strings.TrimSpace(r.Header.Get("Authorization"))
	provided = strings.TrimPrefix(provided, "Bearer ")
	return auth.ValidateToken(provided, token)
}
