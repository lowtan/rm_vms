package logger

import (
	"context"
	"log/slog"
)

// Define a custom type for context keys to prevent collisions
type contextKey string
const TraceIDKey contextKey = "trace_id"

// ContextHandler wraps a standard slog.Handler to inject context values
type ContextHandler struct {
	slog.Handler
}

// Handle intercepts the log record right before it is written as JSON
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Look inside the context for our specific key
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		// If it exists, append it to the log record attributes
		r.AddAttrs(slog.String("trace_id", traceID))
	}

	// Pass the modified record to the underlying JSON handler
	return h.Handler.Handle(ctx, r)
}

// Usage: Setup function to initialize this in your main.go
// func InitLogger() {
// 	// Create the standard JSON handler
// 	baseHandler := slog.NewJSONHandler(os.Stdout, nil)
	
// 	// Wrap it in our custom ContextHandler
// 	slog.SetDefault(slog.New(ContextHandler{baseHandler}))
// }