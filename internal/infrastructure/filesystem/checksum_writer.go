package filesystem

import (
	"encoding/hex"
	"hash"
	"io"
)

// ChecksumWriter wraps a writer and calculates checksum while writing
type ChecksumWriter struct {
	writer io.Writer
	hash   hash.Hash
}

// NewChecksumWriter creates a new ChecksumWriter
func NewChecksumWriter(w io.Writer, h hash.Hash) *ChecksumWriter {
	return &ChecksumWriter{
		writer: w,
		hash:   h,
	}
}

// Write writes data to the underlying writer and updates the hash
func (cw *ChecksumWriter) Write(p []byte) (n int, err error) {
	n, err = cw.writer.Write(p)
	if err != nil {
		return n, err
	}
	// Only hash the bytes that were actually written
	if n > 0 {
		cw.hash.Write(p[:n])
	}
	return n, nil
}

// Sum returns the hexadecimal checksum
func (cw *ChecksumWriter) Sum() string {
	return hex.EncodeToString(cw.hash.Sum(nil))
}

// Close closes the underlying writer if it implements io.Closer
func (cw *ChecksumWriter) Close() error {
	if closer, ok := cw.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
