package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "TEST")

	assert.NotNil(t, log)
	assert.Equal(t, &buf, log.writer)
	assert.Equal(t, "TEST", log.prefix)
}

func TestLogger_Log(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
		format string
		args   []interface{}
		want   string
	}{
		{
			name:   "simple message",
			prefix: "TEST",
			format: "Hello %s",
			args:   []interface{}{"World"},
			want:   "TEST: Hello World",
		},
		{
			name:   "no prefix",
			prefix: "",
			format: "Hello %s",
			args:   []interface{}{"World"},
			want:   "Hello World",
		},
		{
			name:   "with newline",
			prefix: "TEST",
			format: "Hello\n",
			args:   []interface{}{},
			want:   "TEST: Hello\n",
		},
		{
			name:   "no args",
			prefix: "TEST",
			format: "Hello World",
			args:   []interface{}{},
			want:   "TEST: Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(&buf, tt.prefix)

			log.Log(tt.format, tt.args...)

			output := buf.String()
			// Check timestamp format
			assert.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]`, output)
			// Check message content
			assert.Contains(t, output, tt.want)
			// Check newline is added if not present
			assert.True(t, strings.HasSuffix(output, "\n"))
		})
	}
}

func TestLogger_SetWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	log := New(&buf1, "TEST")

	log.Log("Message 1")
	assert.Contains(t, buf1.String(), "Message 1")
	assert.Empty(t, buf2.String())

	log.SetWriter(&buf2)
	log.Log("Message 2")
	assert.Contains(t, buf1.String(), "Message 1")
	assert.NotContains(t, buf1.String(), "Message 2")
	assert.Contains(t, buf2.String(), "Message 2")
}

func TestLogger_SetPrefix(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "OLD")

	log.Log("Message 1")
	assert.Contains(t, buf.String(), "OLD: Message 1")

	buf.Reset()
	log.SetPrefix("NEW")
	log.Log("Message 2")
	assert.Contains(t, buf.String(), "NEW: Message 2")
}

func TestLogger_NilWriter(t *testing.T) {
	log := New(nil, "TEST")
	
	// Should not panic
	assert.NotPanics(t, func() {
		log.Log("This should not panic")
	})
}

func TestLogger_Concurrent(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, "TEST")

	var wg sync.WaitGroup
	numGoroutines := 10
	numMessages := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				log.Log("Message from goroutine %d, iteration %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// Check that all messages were written
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, numGoroutines*numMessages, len(lines))

	// Check that all lines have proper format
	for _, line := range lines {
		assert.Regexp(t, `^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] TEST: Message from goroutine \d+, iteration \d+$`, line)
	}
}