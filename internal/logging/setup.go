package logging

import (
	"log/slog"
	"os"
)

// Setup configures the global slog logger once per process: JSON in production, text elsewhere,
// every line tagged with the service name and, when the context carries one, the request ID.
func Setup(level slog.Level, isProduction bool, serviceName string) {
	opts := &slog.HandlerOptions{Level: level, AddSource: true}
	var handler slog.Handler
	if isProduction {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler).With("service", serviceName))

	// SetDefault also routes the std log package through the handler, at Info by default. Those lines
	// are lifecycle events and failures, so never let the level filter drop them.
	slog.SetLogLoggerLevel(level)
}
