package kubernetes

import (
	"reflect"
	"testing"

	"github.com/cagojeiger/cli-restore/internal/runner"
)

func TestGetNamespaces(t *testing.T) {
	tests := []struct {
		name    string
		runner  *mockRunner
		want    []string
		wantErr bool
	}{
		{
			name: "successful namespace retrieval",
			runner: &mockRunner{
				output: `{
					"items": [
						{"metadata": {"name": "default"}},
						{"metadata": {"name": "kube-system"}},
						{"metadata": {"name": "kube-public"}}
					]
				}`,
				err: nil,
			},
			want:    []string{"default", "kube-system", "kube-public"},
			wantErr: false,
		},
		{
			name: "empty namespace list",
			runner: &mockRunner{
				output: `{"items": []}`,
				err:    nil,
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "kubectl command error",
			runner: &mockRunner{
				output: "",
				err:    runner.ErrTest,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid JSON response",
			runner: &mockRunner{
				output: "invalid json",
				err:    nil,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNamespaces(tt.runner)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetNamespaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPods(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		runner    *mockRunner
		want      []Pod
		wantErr   bool
	}{
		{
			name:      "successful pod retrieval",
			namespace: "default",
			runner: &mockRunner{
				output: `{
					"items": [
						{
							"metadata": {"name": "test-pod-1", "namespace": "default"},
							"status": {
								"phase": "Running",
								"containerStatuses": [
									{"ready": true},
									{"ready": true}
								]
							}
						},
						{
							"metadata": {"name": "test-pod-2", "namespace": "default"},
							"status": {
								"phase": "Pending",
								"containerStatuses": [
									{"ready": false}
								]
							}
						}
					]
				}`,
				err: nil,
			},
			want: []Pod{
				{Name: "test-pod-1", Namespace: "default", Status: "Running", Ready: "2/2"},
				{Name: "test-pod-2", Namespace: "default", Status: "Pending", Ready: "0/1"},
			},
			wantErr: false,
		},
		{
			name:      "empty pod list",
			namespace: "empty",
			runner: &mockRunner{
				output: `{"items": []}`,
				err:    nil,
			},
			want:    []Pod{},
			wantErr: false,
		},
		{
			name:      "kubectl command error",
			namespace: "error",
			runner: &mockRunner{
				output: "",
				err:    runner.ErrTest,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPods(tt.runner, tt.namespace)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPods() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDirectoryContents(t *testing.T) {
	tests := []struct {
		name      string
		runner    *mockRunner
		pod       string
		namespace string
		path      string
		want      []DirectoryEntry
		wantErr   bool
	}{
		{
			name:      "successful directory listing",
			runner: &mockRunner{
				output: `total 12
drwxr-xr-x    3 root     root          4096 Jan  1 12:00 .
drwxr-xr-x   17 root     root          4096 Jan  1 12:00 ..
drwxr-xr-x    2 root     root          4096 Jan  1 12:00 logs
-rw-r--r--    1 root     root           100 Jan  1 12:00 config.yaml`,
				err: nil,
			},
			pod:       "test-pod",
			namespace: "default",
			path:      "/app",
			want: []DirectoryEntry{
				{Name: "logs", Type: "dir", Size: "4096", Modified: "Jan 1 12:00"},
				{Name: "config.yaml", Type: "file", Size: "100", Modified: "Jan 1 12:00"},
			},
			wantErr: false,
		},
		{
			name: "empty directory",
			runner: &mockRunner{
				output: `total 0
drwxr-xr-x    2 root     root          4096 Jan  1 12:00 .
drwxr-xr-x   17 root     root          4096 Jan  1 12:00 ..`,
				err: nil,
			},
			pod:       "test-pod",
			namespace: "default",
			path:      "/empty",
			want:      nil,
			wantErr:   false,
		},
		{
			name: "kubectl command error",
			runner: &mockRunner{
				output: "",
				err:    runner.ErrTest,
			},
			pod:       "test-pod",
			namespace: "default",
			path:      "/error",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDirectoryContents(tt.runner, tt.pod, tt.namespace, tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDirectoryContents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Handle nil vs empty slice comparison
			if (got == nil && tt.want != nil) || (got != nil && tt.want == nil) || 
			   (len(got) != len(tt.want)) {
				t.Errorf("GetDirectoryContents() = %v (len=%d, nil=%v), want %v (len=%d, nil=%v)", 
					got, len(got), got == nil, tt.want, len(tt.want), tt.want == nil)
				return
			}
			
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDirectoryContents() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockRunner implements the runner.Runner interface for testing
type mockRunner struct {
	output string
	err    error
}

func (m *mockRunner) Run(cmd string, args ...string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(m.output), nil
}