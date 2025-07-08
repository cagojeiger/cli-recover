package logger

import (
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "debug level",
			level: DebugLevel,
			want:  "DEBUG",
		},
		{
			name:  "info level",
			level: InfoLevel,
			want:  "INFO",
		},
		{
			name:  "warn level",
			level: WarnLevel,
			want:  "WARN",
		},
		{
			name:  "error level",
			level: ErrorLevel,
			want:  "ERROR",
		},
		{
			name:  "fatal level",
			level: FatalLevel,
			want:  "FATAL",
		},
		{
			name:  "unknown level",
			level: Level(999),
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestF(t *testing.T) {
	field := F("key", "value")
	if field.Key != "key" {
		t.Errorf("F() field.Key = %v, want %v", field.Key, "key")
	}
	if field.Value != "value" {
		t.Errorf("F() field.Value = %v, want %v", field.Value, "value")
	}
}
