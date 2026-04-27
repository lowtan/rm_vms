package logger

import "sync"

// LogBuffer stores a fixed number of string entries.
type LogBuffer struct {
	mu       sync.RWMutex
	items    []string
	capacity int
}

func NewBuffer(capacity int) *LogBuffer {
	return &LogBuffer{
		items:    make([]string, 0, capacity),
		capacity: capacity,
	}
}

// Write makes LogBuffer fulfill the io.Writer interface.
func (b *LogBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.items) >= b.capacity {
		b.items = b.items[1:] // Shift left
	}
	b.items = append(b.items, string(p))
	
	return len(p), nil
}

// Snapshot returns a thread-safe copy of the current items.
func (b *LogBuffer) Snapshot() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	copyItems := make([]string, len(b.items))
	copy(copyItems, b.items)
	return copyItems
}

/**
 * Usage
 */

// Initialize the Storage
// logMemory := buffer.NewBuffer(1000)

// // Configure the Logger (Writing to Stdout AND the Buffer)
// multiWriter := io.MultiWriter(os.Stdout, logMemory)
// logger := slog.New(slog.NewJSONHandler(multiWriter, nil))
// slog.SetDefault(logger) // Set as global default for the app

// Initialize the Delivery System (Injecting the Buffer into the API)
// logAPI := api.NewLogHandler(logMemory)

// Start the HTTP Server
// mux := http.NewServeMux()
// mux.Handle("/api/logs", logAPI)