package logger

// LogProvider defines the exact behavior this API needs to do its job.
type LogProvider interface {
	Snapshot() []string
}

// LogHandler processes HTTP requests for logs.
type LogHandler struct {
	provider LogProvider
}

// NewLogHandler creates a handler using dependency injection.
func NewLogHandler(provider LogProvider) *LogHandler {
	return &LogHandler{
		provider: provider,
	}
}
