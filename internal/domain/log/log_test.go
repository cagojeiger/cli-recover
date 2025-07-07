package log_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cagojeiger/cli-recover/internal/domain/log"
)

func TestNewLog(t *testing.T) {
	tests := []struct {
		name      string
		logType   log.Type
		provider  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid backup log",
			logType:   log.TypeBackup,
			provider:  "filesystem",
			wantError: false,
		},
		{
			name:      "Valid restore log",
			logType:   log.TypeRestore,
			provider:  "minio",
			wantError: false,
		},
		{
			name:      "Invalid log type",
			logType:   "invalid",
			provider:  "filesystem",
			wantError: true,
			errorMsg:  "invalid log type",
		},
		{
			name:      "Empty provider",
			logType:   log.TypeBackup,
			provider:  "",
			wantError: true,
			errorMsg:  "provider cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := log.NewLog(tt.logType, tt.provider)

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, l)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, l)
				assert.Equal(t, tt.logType, l.Type)
				assert.Equal(t, tt.provider, l.Provider)
				assert.Equal(t, log.StatusRunning, l.Status)
				assert.NotEmpty(t, l.ID)
				assert.NotZero(t, l.StartTime)
				assert.Nil(t, l.EndTime)
				assert.NotNil(t, l.Metadata)
			}
		})
	}
}

func TestLog_Complete(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	// Complete the log
	l.Complete()

	assert.Equal(t, log.StatusCompleted, l.Status)
	assert.NotNil(t, l.EndTime)
	assert.True(t, l.EndTime.After(l.StartTime))
}

func TestLog_Fail(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	// Fail the log
	reason := "connection timeout"
	l.Fail(reason)

	assert.Equal(t, log.StatusFailed, l.Status)
	assert.NotNil(t, l.EndTime)
	assert.Equal(t, reason, l.Metadata["error"])
}

func TestLog_Duration(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	// Test running log
	time.Sleep(100 * time.Millisecond)
	duration := l.Duration()
	assert.Greater(t, duration, 100*time.Millisecond)

	// Test completed log
	l.Complete()
	completedDuration := l.Duration()
	assert.Greater(t, completedDuration, 100*time.Millisecond)
	
	// Duration should be fixed after completion
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, completedDuration, l.Duration())
}

func TestLog_Filename(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	filename := l.Filename()
	assert.Contains(t, filename, "backup")
	assert.Contains(t, filename, "filesystem")
	assert.Contains(t, filename, l.ID)
	assert.True(t, strings.HasSuffix(filename, ".log"))
}

func TestLog_GenerateLogPath(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	path := l.GenerateLogPath("/var/log/cli-recover")
	
	assert.Contains(t, path, "/var/log/cli-recover")
	assert.Contains(t, path, "backup")
	assert.Contains(t, path, "filesystem")
	assert.Contains(t, path, l.StartTime.Format("2006-01-02"))
	assert.True(t, strings.HasSuffix(path, ".log"))
}

func TestLog_Metadata(t *testing.T) {
	l, err := log.NewLog(log.TypeBackup, "filesystem")
	require.NoError(t, err)

	// Set metadata
	l.SetMetadata("namespace", "default")
	l.SetMetadata("pod", "test-pod")

	// Get metadata
	ns, ok := l.GetMetadata("namespace")
	assert.True(t, ok)
	assert.Equal(t, "default", ns)

	pod, ok := l.GetMetadata("pod")
	assert.True(t, ok)
	assert.Equal(t, "test-pod", pod)

	// Get non-existent metadata
	_, ok = l.GetMetadata("nonexistent")
	assert.False(t, ok)
}

func TestLog_Validate(t *testing.T) {
	tests := []struct {
		name      string
		log       *log.Log
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid log",
			log: &log.Log{
				ID:       "20240107-150405",
				Type:     log.TypeBackup,
				Provider: "filesystem",
			},
			wantError: false,
		},
		{
			name: "Empty ID",
			log: &log.Log{
				ID:       "",
				Type:     log.TypeBackup,
				Provider: "filesystem",
			},
			wantError: true,
			errorMsg:  "ID cannot be empty",
		},
		{
			name: "Empty type",
			log: &log.Log{
				ID:       "20240107-150405",
				Type:     "",
				Provider: "filesystem",
			},
			wantError: true,
			errorMsg:  "type cannot be empty",
		},
		{
			name: "Invalid provider with slash",
			log: &log.Log{
				ID:       "20240107-150405",
				Type:     log.TypeBackup,
				Provider: "file/system",
			},
			wantError: true,
			errorMsg:  "invalid provider name",
		},
		{
			name: "Invalid provider with dots",
			log: &log.Log{
				ID:       "20240107-150405",
				Type:     log.TypeBackup,
				Provider: "../filesystem",
			},
			wantError: true,
			errorMsg:  "invalid provider name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.log.Validate()

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}