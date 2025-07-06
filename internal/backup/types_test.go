package backup

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		backupType Type
		expected bool
	}{
		{"filesystem is valid", TypeFilesystem, true},
		{"minio is valid", TypeMinio, true},
		{"mongodb is valid", TypeMongoDB, true},
		{"invalid type", Type("invalid"), false},
		{"empty type", Type(""), false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.backupType.IsValid())
		})
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Type
		wantErr bool
	}{
		{"parse filesystem", "filesystem", TypeFilesystem, false},
		{"parse minio", "minio", TypeMinio, false},
		{"parse mongodb", "mongodb", TypeMongoDB, false},
		{"parse uppercase", "FILESYSTEM", TypeFilesystem, false},
		{"parse mixed case", "MinIO", TypeMinio, false},
		{"parse invalid", "invalid", "", true},
		{"parse empty", "", "", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBackupSource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		source  BackupSource
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filesystem backup",
			source: BackupSource{
				Type:   TypeFilesystem,
				Pod:    "my-pod",
				Source: "/var/log",
			},
			wantErr: false,
		},
		{
			name: "filesystem with relative path",
			source: BackupSource{
				Type:   TypeFilesystem,
				Pod:    "my-pod",
				Source: "var/log",
			},
			wantErr: true,
			errMsg:  "must be absolute",
		},
		{
			name: "valid minio backup",
			source: BackupSource{
				Type:   TypeMinio,
				Pod:    "minio-pod",
				Source: "bucket/path/to/data",
			},
			wantErr: false,
		},
		{
			name: "minio without slash",
			source: BackupSource{
				Type:   TypeMinio,
				Pod:    "minio-pod",
				Source: "bucket",
			},
			wantErr: true,
			errMsg:  "must be in format: bucket/path",
		},
		{
			name: "valid mongodb backup",
			source: BackupSource{
				Type:   TypeMongoDB,
				Pod:    "mongo-pod",
				Source: "mydb.users",
			},
			wantErr: false,
		},
		{
			name: "mongodb without dot",
			source: BackupSource{
				Type:   TypeMongoDB,
				Pod:    "mongo-pod",
				Source: "mydb",
			},
			wantErr: true,
			errMsg:  "must be in format: database.collection",
		},
		{
			name: "missing pod name",
			source: BackupSource{
				Type:   TypeFilesystem,
				Pod:    "",
				Source: "/var/log",
			},
			wantErr: true,
			errMsg:  "pod name is required",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   Type
	}{
		{"infer filesystem from absolute path", "/var/log", TypeFilesystem},
		{"infer filesystem from root", "/", TypeFilesystem},
		{"infer minio from bucket path", "bucket/path", TypeMinio},
		{"infer minio from deep path", "bucket/path/to/data", TypeMinio},
		{"infer mongodb from db.collection", "mydb.users", TypeMongoDB},
		{"infer mongodb from complex name", "production.user_accounts", TypeMongoDB},
		{"prefer filesystem for ambiguous", "/data/file.txt", TypeFilesystem},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferType(tt.source)
			assert.Equal(t, tt.want, got)
		})
	}
}