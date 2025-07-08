package backup_test

import (
	"testing"
	"time"

	"github.com/cagojeiger/cli-recover/internal/domain/backup"
	"github.com/stretchr/testify/assert"
)

func TestProgress_CalculateSpeed(t *testing.T) {
	tests := []struct {
		name     string
		progress backup.Progress
		duration time.Duration
		want     float64
	}{
		{
			name: "1MB in 1 second",
			progress: backup.Progress{
				Current: 1048576, // 1MB in bytes
				Total:   10485760,
			},
			duration: time.Second,
			want:     1048576.0,
		},
		{
			name: "2MB in 2 seconds",
			progress: backup.Progress{
				Current: 2097152, // 2MB
				Total:   10485760,
			},
			duration: 2 * time.Second,
			want:     1048576.0, // 1MB/s
		},
		{
			name: "0 bytes in 1 second",
			progress: backup.Progress{
				Current: 0,
				Total:   10485760,
			},
			duration: time.Second,
			want:     0.0,
		},
		{
			name: "zero duration",
			progress: backup.Progress{
				Current: 1048576,
				Total:   10485760,
			},
			duration: 0,
			want:     0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.progress.CalculateSpeed(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgress_CalculateETA(t *testing.T) {
	tests := []struct {
		name     string
		progress backup.Progress
		speed    float64
		want     time.Duration
	}{
		{
			name: "50% complete at 1MB/s",
			progress: backup.Progress{
				Current: 5242880,  // 5MB
				Total:   10485760, // 10MB
			},
			speed: 1048576.0, // 1MB/s
			want:  5 * time.Second,
		},
		{
			name: "25% complete at 2MB/s",
			progress: backup.Progress{
				Current: 2621440,  // 2.5MB
				Total:   10485760, // 10MB
			},
			speed: 2097152.0, // 2MB/s
			want:  time.Duration(3750) * time.Millisecond,
		},
		{
			name: "zero speed",
			progress: backup.Progress{
				Current: 5242880,
				Total:   10485760,
			},
			speed: 0,
			want:  0,
		},
		{
			name: "already complete",
			progress: backup.Progress{
				Current: 10485760,
				Total:   10485760,
			},
			speed: 1048576.0,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.progress.CalculateETA(tt.speed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgress_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		progress backup.Progress
		want     float64
	}{
		{
			name: "0% complete",
			progress: backup.Progress{
				Current: 0,
				Total:   100,
			},
			want: 0.0,
		},
		{
			name: "50% complete",
			progress: backup.Progress{
				Current: 50,
				Total:   100,
			},
			want: 50.0,
		},
		{
			name: "100% complete",
			progress: backup.Progress{
				Current: 100,
				Total:   100,
			},
			want: 100.0,
		},
		{
			name: "zero total",
			progress: backup.Progress{
				Current: 50,
				Total:   0,
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.progress.Percentage()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProgress_FormatETA(t *testing.T) {
	tests := []struct {
		name string
		eta  time.Duration
		want string
	}{
		{
			name: "less than a minute",
			eta:  30 * time.Second,
			want: "30s",
		},
		{
			name: "exactly 1 minute",
			eta:  60 * time.Second,
			want: "1m0s",
		},
		{
			name: "1 minute 30 seconds",
			eta:  90 * time.Second,
			want: "1m30s",
		},
		{
			name: "over an hour",
			eta:  3665 * time.Second,
			want: "1h1m5s",
		},
		{
			name: "zero duration",
			eta:  0,
			want: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backup.FormatETA(tt.eta)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOptions_Validate_Success(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		options := backup.Options{
			Namespace:  "default",
			PodName:    "my-pod",
			SourcePath: "/data",
			OutputFile: "backup.tar.gz",
		}

		err := options.Validate()

		assert.NoError(t, err)
	})
}

func TestOptions_Validate_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		options backup.Options
		errMsg  string
	}{
		{
			name: "missing namespace",
			options: backup.Options{
				PodName:    "my-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar.gz",
			},
			errMsg: "namespace is required",
		},
		{
			name: "missing pod name",
			options: backup.Options{
				Namespace:  "default",
				SourcePath: "/data",
				OutputFile: "backup.tar.gz",
			},
			errMsg: "pod name is required",
		},
		{
			name: "missing source path",
			options: backup.Options{
				Namespace:  "default",
				PodName:    "my-pod",
				OutputFile: "backup.tar.gz",
			},
			errMsg: "source path is required",
		},
		{
			name: "missing output file",
			options: backup.Options{
				Namespace:  "default",
				PodName:    "my-pod",
				SourcePath: "/data",
			},
			errMsg: "output file is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestBackupError_Wrap(t *testing.T) {
	originalErr := assert.AnError

	err := backup.NewBackupError("operation failed", originalErr)

	assert.Equal(t, "operation failed", err.Message)
	assert.Equal(t, originalErr, err.Cause)
	assert.NotZero(t, err.Timestamp)
}

func TestBackupError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  backup.BackupError
		want string
	}{
		{
			name: "error without cause",
			err: backup.BackupError{
				Code:    backup.ErrCodeInternal,
				Message: "internal error",
			},
			want: "[INTERNAL] internal error",
		},
		{
			name: "error with cause",
			err: backup.BackupError{
				Code:    backup.ErrCodeNotFound,
				Message: "resource not found",
				Cause:   assert.AnError,
			},
			want: "[NOT_FOUND] resource not found: assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBackupError_Is(t *testing.T) {
	err1 := backup.BackupError{Code: backup.ErrCodeNotFound, Message: "not found"}
	err2 := backup.BackupError{Code: backup.ErrCodeNotFound, Message: "also not found"}
	err3 := backup.BackupError{Code: backup.ErrCodeInternal, Message: "internal"}

	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(assert.AnError))
}
