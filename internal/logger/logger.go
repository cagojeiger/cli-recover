package logger

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Logger provides thread-safe logging functionality
type Logger struct {
	mu     sync.Mutex
	writer io.Writer
	prefix string
}

// New creates a new logger
func New(writer io.Writer, prefix string) *Logger {
	return &Logger{
		writer: writer,
		prefix: prefix,
	}
}

// Log writes a log message
func (l *Logger) Log(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	
	if l.prefix != "" {
		fmt.Fprintf(l.writer, "[%s] %s: %s", timestamp, l.prefix, message)
	} else {
		fmt.Fprintf(l.writer, "[%s] %s", timestamp, message)
	}

	// Add newline if not present
	if len(message) > 0 && message[len(message)-1] != '\n' {
		fmt.Fprint(l.writer, "\n")
	}
}

// SetWriter changes the log writer
func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer = w
}

// SetPrefix changes the log prefix
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}