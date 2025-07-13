package valueobject

import (
	"testing"
)

func TestStream_NewStreamName(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid stream name",
			streamName: "test-stream",
			wantErr:    false,
		},
		{
			name:       "valid with underscore",
			streamName: "test_stream_123",
			wantErr:    false,
		},
		{
			name:       "empty name should fail",
			streamName: "",
			wantErr:    true,
			errMsg:     "stream name cannot be empty",
		},
		{
			name:       "name with spaces should fail",
			streamName: "test stream",
			wantErr:    true,
			errMsg:     "stream name contains invalid characters",
		},
		{
			name:       "name with special chars should fail",
			streamName: "test@stream",
			wantErr:    true,
			errMsg:     "stream name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream, err := NewStreamName(tt.streamName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewStreamName() error = nil, wantErr %v", tt.wantErr)
				} else if err.Error() != tt.errMsg {
					t.Errorf("NewStreamName() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewStreamName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if stream.String() != tt.streamName {
				t.Errorf("StreamName.String() = %v, want %v", stream.String(), tt.streamName)
			}
		})
	}
}

func TestStreamType_String(t *testing.T) {
	tests := []struct {
		name string
		st   StreamType
		want string
	}{
		{
			name: "stream type",
			st:   StreamTypeStream,
			want: "stream",
		},
		{
			name: "file type",
			st:   StreamTypeFile,
			want: "file",
		},
		{
			name: "variable type",
			st:   StreamTypeVariable,
			want: "variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.String(); got != tt.want {
				t.Errorf("StreamType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamReference_Parse(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		wantType StreamType
		wantName string
		wantErr  bool
	}{
		{
			name:     "simple stream reference",
			ref:      "my-stream",
			wantType: StreamTypeStream,
			wantName: "my-stream",
			wantErr:  false,
		},
		{
			name:     "file reference",
			ref:      "file:output.txt",
			wantType: StreamTypeFile,
			wantName: "output.txt",
			wantErr:  false,
		},
		{
			name:     "variable reference",
			ref:      "var:total_size",
			wantType: StreamTypeVariable,
			wantName: "total_size",
			wantErr:  false,
		},
		{
			name:     "empty reference should fail",
			ref:      "",
			wantType: StreamTypeStream,
			wantName: "",
			wantErr:  true,
		},
		{
			name:     "invalid prefix should fail",
			ref:      "invalid:something",
			wantType: StreamTypeStream,
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr, err := ParseStreamReference(tt.ref)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseStreamReference() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseStreamReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if sr.Type != tt.wantType {
				t.Errorf("StreamReference.Type = %v, want %v", sr.Type, tt.wantType)
			}

			if sr.Name.String() != tt.wantName {
				t.Errorf("StreamReference.Name = %v, want %v", sr.Name.String(), tt.wantName)
			}
		})
	}
}