package pipeline

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// ChecksumWriter 는 데이터를 받으면서 체크섬을 계산
type ChecksumWriter struct {
	algorithm string
	hash      hash.Hash
	mu        sync.Mutex
}

// NewChecksumWriter 는 새로운 ChecksumWriter를 생성
func NewChecksumWriter(algorithm string) *ChecksumWriter {
	var h hash.Hash
	
	switch algorithm {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		// 기본값은 sha256
		h = sha256.New()
	}
	
	return &ChecksumWriter{
		algorithm: algorithm,
		hash:      h,
	}
}

// Write 는 데이터를 받아 체크섬을 업데이트
func (w *ChecksumWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	return w.hash.Write(p)
}

// Update 는 Monitor 인터페이스 구현 (Write와 동일)
func (w *ChecksumWriter) Update(bytes int64) {
	// ChecksumWriter는 Write를 통해 데이터를 받으므로
	// Update는 사용하지 않음
}

// Finish 는 Monitor 인터페이스 구현
func (w *ChecksumWriter) Finish() {
	// ChecksumWriter는 특별한 종료 처리가 필요 없음
}

// Report 는 Monitor 인터페이스 구현
func (w *ChecksumWriter) Report() string {
	return fmt.Sprintf("Checksum (%s): %s", w.algorithm, w.Sum())
}

// Sum 은 현재까지 계산된 체크섬을 16진수 문자열로 반환
func (w *ChecksumWriter) Sum() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	return hex.EncodeToString(w.hash.Sum(nil))
}

// Algorithm 은 사용 중인 알고리즘을 반환
func (w *ChecksumWriter) Algorithm() string {
	return w.algorithm
}

// ChecksumFileWriter 는 체크섬을 계산하고 파일로 저장
type ChecksumFileWriter struct {
	*ChecksumWriter
	outputPath string
}

// NewChecksumFileWriter 는 체크섬을 파일로 저장하는 writer 생성
func NewChecksumFileWriter(algorithm, outputPath string) *ChecksumFileWriter {
	return &ChecksumFileWriter{
		ChecksumWriter: NewChecksumWriter(algorithm),
		outputPath:     outputPath,
	}
}

// SaveToFile 은 계산된 체크섬을 파일로 저장
func (w *ChecksumFileWriter) SaveToFile() error {
	// 체크섬 파일 경로 생성
	checksumPath := w.outputPath + "." + w.algorithm
	if w.algorithm == "unknown" {
		checksumPath = w.outputPath + ".sha256"
	}
	
	// 체크섬 내용 생성 (GNU coreutils 형식)
	content := fmt.Sprintf("%s  %s\n", w.Sum(), filepath.Base(w.outputPath))
	
	// 원자적 파일 쓰기 (임시 파일 -> 이름 변경)
	tempPath := checksumPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// 임시 파일을 최종 파일로 이동
	if err := os.Rename(tempPath, checksumPath); err != nil {
		os.Remove(tempPath) // 실패 시 임시 파일 정리
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	return nil
}

// MultiChecksumWriter 는 여러 체크섬을 동시에 계산
type MultiChecksumWriter struct {
	writers map[string]*ChecksumWriter
	mu      sync.Mutex
}

// NewMultiChecksumWriter 는 여러 알고리즘의 체크섬을 동시에 계산하는 writer 생성
func NewMultiChecksumWriter(algorithms []string) *MultiChecksumWriter {
	writers := make(map[string]*ChecksumWriter)
	
	for _, algo := range algorithms {
		writers[algo] = NewChecksumWriter(algo)
	}
	
	return &MultiChecksumWriter{
		writers: writers,
	}
}

// Write 는 모든 체크섬 writer에 데이터를 전달
func (w *MultiChecksumWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	for _, writer := range w.writers {
		if _, err := writer.Write(p); err != nil {
			return 0, err
		}
	}
	
	return len(p), nil
}

// Sums 는 모든 알고리즘의 체크섬을 반환
func (w *MultiChecksumWriter) Sums() map[string]string {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	results := make(map[string]string)
	for algo, writer := range w.writers {
		results[algo] = writer.Sum()
	}
	
	return results
}

// ChecksumVerifier 는 파일의 체크섬을 검증
type ChecksumVerifier struct {
	algorithm string
}

// NewChecksumVerifier 는 체크섬 검증기를 생성
func NewChecksumVerifier(algorithm string) *ChecksumVerifier {
	return &ChecksumVerifier{
		algorithm: algorithm,
	}
}

// VerifyFile 은 파일의 체크섬이 예상값과 일치하는지 확인
func (v *ChecksumVerifier) VerifyFile(filePath, expectedSum string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	cw := NewChecksumWriter(v.algorithm)
	if _, err := io.Copy(cw, file); err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	
	actualSum := cw.Sum()
	return actualSum == expectedSum, nil
}