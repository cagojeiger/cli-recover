package adapters

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cagojeiger/cli-recover/internal/domain/restore"
)

// MockMetadataStore is a mock implementation of metadata.Store
type MockMetadataStore struct {
	mock.Mock
}

func (m *MockMetadataStore) Save(metadata *restore.Metadata) error {
	args := m.Called(metadata)
	return args.Error(0)
}

func (m *MockMetadataStore) Get(id string) (*restore.Metadata, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*restore.Metadata), args.Error(1)
}

func (m *MockMetadataStore) GetByFile(backupFile string) (*restore.Metadata, error) {
	args := m.Called(backupFile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*restore.Metadata), args.Error(1)
}

func (m *MockMetadataStore) List() ([]*restore.Metadata, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*restore.Metadata), args.Error(1)
}

func (m *MockMetadataStore) ListByNamespace(namespace string) ([]*restore.Metadata, error) {
	args := m.Called(namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*restore.Metadata), args.Error(1)
}

func (m *MockMetadataStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestListAdapter_ExecuteList(t *testing.T) {
	// Sample metadata for testing
	now := time.Now()
	sampleMetadata := []*restore.Metadata{
		{
			ID:          "backup-1",
			Type:        "filesystem",
			Namespace:   "default",
			PodName:     "app-pod",
			SourcePath:  "/data",
			BackupFile:  "backup-default-app-pod-data.tar.gz",
			Size:        1024 * 1024 * 50, // 50MB
			CreatedAt:   now.Add(-2 * time.Hour),
			CompletedAt: now.Add(-2 * time.Hour).Add(5 * time.Minute),
			Status:      "completed",
			Compression: "gzip",
		},
		{
			ID:          "backup-2",
			Type:        "filesystem",
			Namespace:   "production",
			PodName:     "db-pod",
			SourcePath:  "/var/lib/data",
			BackupFile:  "backup-production-db-pod-var-lib-data.tar.gz",
			Size:        1024 * 1024 * 150, // 150MB
			CreatedAt:   now.Add(-1 * time.Hour),
			CompletedAt: now.Add(-1 * time.Hour).Add(10 * time.Minute),
			Status:      "completed",
			Compression: "gzip",
		},
	}

	tests := []struct {
		name           string
		setupFlags     func(*cobra.Command)
		setupMock      func(*MockMetadataStore)
		wantErr        bool
		expectedOutput []string
		notExpected    []string
	}{
		{
			name: "list all backups",
			setupFlags: func(cmd *cobra.Command) {
				// No specific flags
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return(sampleMetadata, nil)
			},
			expectedOutput: []string{
				"ID", "Type", "Namespace", "Pod", "Path", "Size", "Created",
				"backup-1", "filesystem", "default", "app-pod", "/data", "50.0 MB",
				"backup-2", "filesystem", "production", "db-pod", "/var/lib/data", "150.0 MB",
				"Total: 2 backups",
			},
		},
		{
			name: "list backups by namespace",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("namespace", "production")
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("ListByNamespace", "production").Return([]*restore.Metadata{sampleMetadata[1]}, nil)
			},
			expectedOutput: []string{
				"backup-2", "production", "db-pod",
				"Total: 1 backup",
			},
			notExpected: []string{
				"backup-1", "default", "app-pod",
			},
		},
		{
			name: "list with output format json",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("output", "json")
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return(sampleMetadata[:1], nil) // Just one for simpler test
			},
			expectedOutput: []string{
				`"id": "backup-1"`,
				`"type": "filesystem"`,
				`"namespace": "default"`,
				`"pod_name": "app-pod"`,
			},
		},
		{
			name: "list with output format yaml",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("output", "yaml")
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return(sampleMetadata[:1], nil)
			},
			expectedOutput: []string{
				"id: backup-1",
				"type: filesystem",
				"namespace: default",
				"podname: app-pod",
			},
		},
		{
			name: "list with details flag",
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Set("details", "true")
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return(sampleMetadata[:1], nil)
			},
			expectedOutput: []string{
				"Backup ID:", "backup-1",
				"Type:", "filesystem",
				"Namespace:", "default",
				"Pod:", "app-pod",
				"Source Path:", "/data",
				"Backup File:", "backup-default-app-pod-data.tar.gz",
				"Size:", "50.0 MB",
				"Compression:", "gzip",
				"Status:", "completed",
				"Created:",
				"Completed:",
				"Duration:",
			},
		},
		{
			name: "empty backup list",
			setupFlags: func(cmd *cobra.Command) {
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return([]*restore.Metadata{}, nil)
			},
			expectedOutput: []string{
				"No backups found",
			},
		},
		{
			name: "error retrieving backups",
			setupFlags: func(cmd *cobra.Command) {
			},
			setupMock: func(m *MockMetadataStore) {
				m.On("List").Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test command
			cmd := &cobra.Command{}
			cmd.Flags().String("namespace", "", "Filter by namespace")
			cmd.Flags().String("output", "table", "Output format (table, json, yaml)")
			cmd.Flags().Bool("details", false, "Show detailed information")
			
			if tt.setupFlags != nil {
				tt.setupFlags(cmd)
			}

			// Create mock store
			mockStore := new(MockMetadataStore)
			if tt.setupMock != nil {
				tt.setupMock(mockStore)
			}

			// Create adapter
			adapter := &ListAdapter{
				store: mockStore,
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute
			err := adapter.ExecuteList(cmd, []string{})

			// Restore output
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check expected output
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}

			// Check not expected
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected)
			}

			// Verify mock expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestListAdapter_FormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024*1024*1024 + 512*1024*1024, "1.5 GB"},
	}

	adapter := &ListAdapter{}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := adapter.formatSize(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListAdapter_FormatDuration(t *testing.T) {
	tests := []struct {
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			start:    time.Now(),
			end:      time.Now().Add(30 * time.Second),
			expected: "30s",
		},
		{
			start:    time.Now(),
			end:      time.Now().Add(5 * time.Minute),
			expected: "5m0s",
		},
		{
			start:    time.Now(),
			end:      time.Now().Add(1*time.Hour + 30*time.Minute),
			expected: "1h30m0s",
		},
	}

	adapter := &ListAdapter{}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := adapter.formatDuration(tt.start, tt.end)
			assert.Equal(t, tt.expected, result)
		})
	}
}