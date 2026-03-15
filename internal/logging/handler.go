package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
// TODO(v0.3.0): Add authentication for log-level endpoint
func NewLogLevelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

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
