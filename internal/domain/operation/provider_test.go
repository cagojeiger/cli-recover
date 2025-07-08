package operation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of the Provider interface
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Type() ProviderType {
	args := m.Called()
	return args.Get(0).(ProviderType)
}

func (m *MockProvider) ValidateOptions(opts Options) error {
	args := m.Called(opts)
	return args.Error(0)
}

func (m *MockProvider) Execute(ctx context.Context, opts Options) (*Result, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Result), args.Error(1)
}

func (m *MockProvider) EstimateSize(opts Options) (int64, error) {
	args := m.Called(opts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProvider) StreamProgress() <-chan Progress {
	args := m.Called()
	return args.Get(0).(<-chan Progress)
}

// TestProviderInterface tests the unified Provider interface
func TestProviderInterface(t *testing.T) {
	tests := []struct {
		name     string
		provider func() *MockProvider
		testFunc func(t *testing.T, p Provider)
	}{
		{
			name: "backup provider basic operations",
			provider: func() *MockProvider {
				m := new(MockProvider)
				m.On("Name").Return("filesystem")
				m.On("Description").Return("Filesystem backup provider")
				m.On("Type").Return(TypeBackup)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				assert.Equal(t, "filesystem", p.Name())
				assert.Equal(t, "Filesystem backup provider", p.Description())
				assert.Equal(t, TypeBackup, p.Type())
			},
		},
		{
			name: "restore provider basic operations",
			provider: func() *MockProvider {
				m := new(MockProvider)
				m.On("Name").Return("filesystem")
				m.On("Description").Return("Filesystem restore provider")
				m.On("Type").Return(TypeRestore)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				assert.Equal(t, "filesystem", p.Name())
				assert.Equal(t, "Filesystem restore provider", p.Description())
				assert.Equal(t, TypeRestore, p.Type())
			},
		},
		{
			name: "validate options",
			provider: func() *MockProvider {
				m := new(MockProvider)
				opts := Options{
					Namespace:  "default",
					PodName:    "test-pod",
					SourcePath: "/data",
				}
				m.On("ValidateOptions", opts).Return(nil)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				opts := Options{
					Namespace:  "default",
					PodName:    "test-pod",
					SourcePath: "/data",
				}
				err := p.ValidateOptions(opts)
				assert.NoError(t, err)
			},
		},
		{
			name: "execute backup operation",
			provider: func() *MockProvider {
				m := new(MockProvider)
				result := &Result{
					Success:      true,
					Message:      "Backup completed",
					BytesWritten: 1024,
					FileCount:    10,
				}
				m.On("Execute", mock.Anything, mock.Anything).Return(result, nil)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				ctx := context.Background()
				opts := Options{
					Namespace:  "default",
					PodName:    "test-pod",
					SourcePath: "/data",
				}
				result, err := p.Execute(ctx, opts)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, "Backup completed", result.Message)
				assert.Equal(t, int64(1024), result.BytesWritten)
				assert.Equal(t, 10, result.FileCount)
			},
		},
		{
			name: "execute restore operation",
			provider: func() *MockProvider {
				m := new(MockProvider)
				result := &Result{
					Success:      true,
					Message:      "Restore completed",
					RestoredPath: "/data",
					FileCount:    15,
					BytesWritten: 2048,
				}
				m.On("Execute", mock.Anything, mock.Anything).Return(result, nil)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				ctx := context.Background()
				opts := Options{
					Namespace:   "default",
					PodName:     "test-pod",
					BackupFile:  "backup.tar",
					TargetPath:  "/data",
				}
				result, err := p.Execute(ctx, opts)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, "Restore completed", result.Message)
				assert.Equal(t, "/data", result.RestoredPath)
				assert.Equal(t, 15, result.FileCount)
				assert.Equal(t, int64(2048), result.BytesWritten)
			},
		},
		{
			name: "estimate size",
			provider: func() *MockProvider {
				m := new(MockProvider)
				m.On("EstimateSize", mock.Anything).Return(int64(1024*1024), nil)
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				opts := Options{
					SourcePath: "/data",
				}
				size, err := p.EstimateSize(opts)
				assert.NoError(t, err)
				assert.Equal(t, int64(1024*1024), size)
			},
		},
		{
			name: "stream progress",
			provider: func() *MockProvider {
				m := new(MockProvider)
				ch := make(chan Progress, 2)
				ch <- Progress{Current: 100, Total: 1000, Message: "Processing..."}
				ch <- Progress{Current: 1000, Total: 1000, Message: "Complete"}
				close(ch)
				m.On("StreamProgress").Return((<-chan Progress)(ch))
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				progressCh := p.StreamProgress()
				
				var updates []Progress
				for progress := range progressCh {
					updates = append(updates, progress)
				}
				
				assert.Len(t, updates, 2)
				assert.Equal(t, int64(100), updates[0].Current)
				assert.Equal(t, int64(1000), updates[1].Current)
			},
		},
		{
			name: "error handling",
			provider: func() *MockProvider {
				m := new(MockProvider)
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, errors.New("operation failed"))
				return m
			},
			testFunc: func(t *testing.T, p Provider) {
				ctx := context.Background()
				opts := Options{}
				result, err := p.Execute(ctx, opts)
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "operation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := tt.provider()
			tt.testFunc(t, provider)
			provider.AssertExpectations(t)
		})
	}
}

// TestUnifiedOptions tests the unified Options structure
func TestUnifiedOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		validate func(t *testing.T, opts Options)
	}{
		{
			name: "backup options",
			opts: Options{
				Type:       TypeBackup,
				Namespace:  "default",
				PodName:    "test-pod",
				SourcePath: "/data",
				OutputFile: "backup.tar.gz",
				Compress:   true,
				Exclude:    []string{"*.log", "*.tmp"},
			},
			validate: func(t *testing.T, opts Options) {
				assert.Equal(t, TypeBackup, opts.Type)
				assert.Equal(t, "default", opts.Namespace)
				assert.Equal(t, "test-pod", opts.PodName)
				assert.Equal(t, "/data", opts.SourcePath)
				assert.Equal(t, "backup.tar.gz", opts.OutputFile)
				assert.True(t, opts.Compress)
				assert.Contains(t, opts.Exclude, "*.log")
			},
		},
		{
			name: "restore options",
			opts: Options{
				Type:          TypeRestore,
				Namespace:     "production",
				PodName:       "app-pod",
				BackupFile:    "backup.tar",
				TargetPath:    "/restore/data",
				Overwrite:     true,
				PreservePerms: true,
			},
			validate: func(t *testing.T, opts Options) {
				assert.Equal(t, TypeRestore, opts.Type)
				assert.Equal(t, "production", opts.Namespace)
				assert.Equal(t, "app-pod", opts.PodName)
				assert.Equal(t, "backup.tar", opts.BackupFile)
				assert.Equal(t, "/restore/data", opts.TargetPath)
				assert.True(t, opts.Overwrite)
				assert.True(t, opts.PreservePerms)
			},
		},
		{
			name: "options with extra data",
			opts: Options{
				Type:      TypeBackup,
				Namespace: "default",
				Extra: map[string]interface{}{
					"verbose": true,
					"debug":   false,
					"timeout": 300,
				},
			},
			validate: func(t *testing.T, opts Options) {
				assert.Equal(t, TypeBackup, opts.Type)
				assert.True(t, opts.Extra["verbose"].(bool))
				assert.False(t, opts.Extra["debug"].(bool))
				assert.Equal(t, 300, opts.Extra["timeout"].(int))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.opts)
		})
	}
}

// TestUnifiedResult tests the unified Result structure
func TestUnifiedResult(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		validate func(t *testing.T, r Result)
	}{
		{
			name: "backup result",
			result: Result{
				Success:      true,
				Message:      "Backup completed successfully",
				BytesWritten: 1024 * 1024,
				FileCount:    100,
				Warnings:     []string{"Skipped .git directory"},
			},
			validate: func(t *testing.T, r Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "Backup completed successfully", r.Message)
				assert.Equal(t, int64(1024*1024), r.BytesWritten)
				assert.Equal(t, 100, r.FileCount)
				assert.Len(t, r.Warnings, 1)
			},
		},
		{
			name: "restore result",
			result: Result{
				Success:      true,
				Message:      "Restore completed successfully",
				RestoredPath: "/data",
				FileCount:    50,
				BytesWritten: 512 * 1024,
				Warnings:     []string{"File permissions not preserved for 2 files"},
			},
			validate: func(t *testing.T, r Result) {
				assert.True(t, r.Success)
				assert.Equal(t, "Restore completed successfully", r.Message)
				assert.Equal(t, "/data", r.RestoredPath)
				assert.Equal(t, 50, r.FileCount)
				assert.Equal(t, int64(512*1024), r.BytesWritten)
				assert.Contains(t, r.Warnings[0], "permissions")
			},
		},
		{
			name: "failed result",
			result: Result{
				Success: false,
				Message: "Operation failed: disk full",
				Error:   errors.New("no space left on device"),
			},
			validate: func(t *testing.T, r Result) {
				assert.False(t, r.Success)
				assert.Contains(t, r.Message, "disk full")
				assert.Error(t, r.Error)
				assert.Contains(t, r.Error.Error(), "no space left")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.result)
		})
	}
}

// TestProviderRegistry tests the provider registry functionality
func TestProviderRegistry(t *testing.T) {
	t.Run("register and retrieve providers", func(t *testing.T) {
		registry := NewProviderRegistry()
		
		// Create mock providers
		backupProvider := new(MockProvider)
		backupProvider.On("Name").Return("filesystem")
		backupProvider.On("Type").Return(TypeBackup)
		
		restoreProvider := new(MockProvider)
		restoreProvider.On("Name").Return("filesystem")
		restoreProvider.On("Type").Return(TypeRestore)
		
		// Register providers
		err := registry.Register(backupProvider)
		assert.NoError(t, err)
		
		err = registry.Register(restoreProvider)
		assert.NoError(t, err)
		
		// Get backup provider
		provider, err := registry.Get("filesystem", TypeBackup)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, TypeBackup, provider.Type())
		
		// Get restore provider
		provider, err = registry.Get("filesystem", TypeRestore)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, TypeRestore, provider.Type())
		
		// List providers
		providers := registry.List()
		assert.Len(t, providers, 2)
	})
	
	t.Run("duplicate registration", func(t *testing.T) {
		registry := NewProviderRegistry()
		
		provider1 := new(MockProvider)
		provider1.On("Name").Return("test")
		provider1.On("Type").Return(TypeBackup)
		
		provider2 := new(MockProvider)
		provider2.On("Name").Return("test")
		provider2.On("Type").Return(TypeBackup)
		
		err := registry.Register(provider1)
		assert.NoError(t, err)
		
		err = registry.Register(provider2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
	
	t.Run("provider not found", func(t *testing.T) {
		registry := NewProviderRegistry()
		
		_, err := registry.Get("nonexistent", TypeBackup)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestProgress tests the Progress structure
func TestProgress(t *testing.T) {
	progress := Progress{
		Current: 500,
		Total:   1000,
		Message: "Processing file: data.txt",
	}
	
	assert.Equal(t, int64(500), progress.Current)
	assert.Equal(t, int64(1000), progress.Total)
	assert.Equal(t, "Processing file: data.txt", progress.Message)
	
	// Calculate percentage
	percentage := float64(progress.Current) / float64(progress.Total) * 100
	assert.Equal(t, float64(50), percentage)
}