package progress

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockReporter for testing
type mockReporter struct {
	mu            sync.Mutex
	started       bool
	completed     bool
	errorCalled   bool
	updates       []mockUpdate
	startOp       string
	startTotal    int64
	completeTime  time.Time
	errorReceived error
}

type mockUpdate struct {
	current int64
	total   int64
	time    time.Time
}

func (m *mockReporter) Start(operation string, total int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = true
	m.startOp = operation
	m.startTotal = total
}

func (m *mockReporter) Update(current, total int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updates = append(m.updates, mockUpdate{
		current: current,
		total:   total,
		time:    time.Now(),
	})
}

func (m *mockReporter) Complete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completed = true
	m.completeTime = time.Now()
}

func (m *mockReporter) Error(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorCalled = true
	m.errorReceived = err
}

func TestDelayedReporter_QuickOperation(t *testing.T) {
	// Test operation that completes in less than 3 seconds
	mock := &mockReporter{}
	delayed := NewDelayedReporter(mock)

	delayed.Start("Quick operation", 100)
	delayed.Update(50, 100)
	delayed.Update(100, 100)
	
	// Complete before 3 seconds
	time.Sleep(100 * time.Millisecond)
	delayed.Complete()

	// Give a bit of time to ensure no delayed start happens
	time.Sleep(100 * time.Millisecond)

	// Should not have started progress reporting
	assert.False(t, mock.started, "Progress should not be shown for quick operations")
	assert.False(t, mock.completed, "Complete should not be called for quick operations")
	assert.Empty(t, mock.updates, "No updates should be recorded for quick operations")
}

func TestDelayedReporter_LongOperation(t *testing.T) {
	// Test operation that takes longer than 3 seconds
	mock := &mockReporter{}
	delayed := NewDelayedReporter(mock)
	delayed.delayPeriod = 100 * time.Millisecond // Short delay for testing

	delayed.Start("Long operation", 1000)
	
	// Send updates before delay
	delayed.Update(100, 1000)
	delayed.Update(200, 1000)
	
	// Should not have started yet
	assert.False(t, mock.started, "Progress should not be shown before delay")
	
	// Wait for delay to pass
	time.Sleep(150 * time.Millisecond)
	
	// Now it should have started
	assert.True(t, mock.started, "Progress should be shown after delay")
	assert.Equal(t, "Long operation", mock.startOp)
	assert.Equal(t, int64(1000), mock.startTotal)
	
	// Should have received the latest update
	assert.Len(t, mock.updates, 1, "Should have one update after start")
	assert.Equal(t, int64(200), mock.updates[0].current)
	
	// Send more updates after start
	delayed.Update(500, 1000)
	delayed.Update(1000, 1000)
	
	// Should pass through immediately now
	assert.Len(t, mock.updates, 3, "Should have all updates after start")
	
	delayed.Complete()
	assert.True(t, mock.completed, "Complete should be called")
}

func TestDelayedReporter_Error(t *testing.T) {
	// Test that errors are shown immediately
	mock := &mockReporter{}
	delayed := NewDelayedReporter(mock)

	delayed.Start("Error operation", 100)
	
	// Error before delay
	testErr := errors.New("test error")
	delayed.Error(testErr)
	
	// Error should be shown immediately
	assert.True(t, mock.started, "Should start on error")
	assert.True(t, mock.errorCalled, "Error should be called")
	assert.Equal(t, testErr, mock.errorReceived)
}

func TestDelayedReporter_CompleteAfterDelayBeforeStart(t *testing.T) {
	// Test completing after delay period but before progress was shown
	mock := &mockReporter{}
	delayed := NewDelayedReporter(mock)
	delayed.delayPeriod = 100 * time.Millisecond // Short delay for testing

	delayed.Start("Medium operation", 500)
	delayed.Update(250, 500)
	
	// Wait longer than delay but complete before it triggers
	time.Sleep(150 * time.Millisecond)
	delayed.Complete()
	
	// Should show start and complete
	assert.True(t, mock.started, "Should start when completing after delay")
	assert.True(t, mock.completed, "Should complete")
}

func TestDelayedReporter_MultipleOperations(t *testing.T) {
	// Test starting a new operation cancels the previous timer
	mock := &mockReporter{}
	delayed := NewDelayedReporter(mock)
	delayed.delayPeriod = 200 * time.Millisecond

	// Start first operation
	delayed.Start("First operation", 100)
	time.Sleep(100 * time.Millisecond)
	
	// Start second operation before first delay completes
	delayed.Start("Second operation", 200)
	time.Sleep(250 * time.Millisecond)
	
	// Only second operation should have started
	assert.True(t, mock.started)
	assert.Equal(t, "Second operation", mock.startOp)
	assert.Equal(t, int64(200), mock.startTotal)
}