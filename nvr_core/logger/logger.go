package logger

import (
	"fmt"
	"log"
	"os"
)

// Logger wraps the standard Go logger to provide chainable prefixes.
type Logger struct {
	prefix string
	base   *log.Logger
}

// WithPrefix creates a new base logger.
func WithPrefix(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
		// log.New ensures thread-safe, non-interleaved writes to stdout.
		// You can remove log.Ldate|log.Ltime if you only want the raw text.
		base: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}

func (l *Logger) WithPrefixf(prefix string, v ...any) *Logger {
	formatted := fmt.Sprintf(prefix, v...)
	return &Logger{
		prefix: formatted,
		base: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
}


// Lin spawns a child logger that inherits and appends to the parent's prefix.
func (l *Logger) Lin(subPrefix string) *Logger {
	return &Logger{
		prefix: l.prefix + subPrefix,
		base:   l.base, // Share the underlying thread-safe writer
	}
}

func (l *Logger) Linf(subPrefix string, v ...any) *Logger {
	formatted := fmt.Sprintf(subPrefix, v...)
	return &Logger{
		prefix: l.prefix + formatted,
		base:   l.base, // Share the underlying thread-safe writer
	}
}

// Log acts like fmt.Println, automatically handling the prefix spacing.
func (l *Logger) Log(v ...any) {
	message := fmt.Sprint(v...)
	l.base.Printf("%s %s", l.prefix, message)
}

// Logf provides formatted logging (like fmt.Printf).
func (l *Logger) Logf(format string, v ...any) {
	message := fmt.Sprintf(format, v...)
	l.base.Printf("%s %s", l.prefix, message)
}