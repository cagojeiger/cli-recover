package pipeline

import (
	"fmt"
	"sync"
	"time"
)

// Monitor 는 파이프라인 실행 중 메트릭을 추적하는 인터페이스
type Monitor interface {
	Update(bytes int64)
	Finish()
	Report() string
}

// ByteMonitor 는 처리된 바이트 수를 추적
type ByteMonitor struct {
	total int64
	mu    sync.Mutex
}

// NewByteMonitor 는 새로운 ByteMonitor를 생성
func NewByteMonitor() *ByteMonitor {
	return &ByteMonitor{}
}

// Update 는 처리된 바이트 수를 업데이트
func (m *ByteMonitor) Update(bytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.total += bytes
}

// Total 은 총 처리된 바이트 수를 반환
func (m *ByteMonitor) Total() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.total
}

// Finish 는 모니터링을 종료
func (m *ByteMonitor) Finish() {
	// ByteMonitor는 특별한 종료 처리가 필요 없음
}

// Report 는 처리 결과를 문자열로 반환
func (m *ByteMonitor) Report() string {
	total := m.Total()
	return fmt.Sprintf("Processed %d bytes (%s)", total, humanizeBytes(total))
}

// LineMonitor 는 처리된 라인 수를 추적
type LineMonitor struct {
	lines int64
	mu    sync.Mutex
}

// NewLineMonitor 는 새로운 LineMonitor를 생성
func NewLineMonitor() *LineMonitor {
	return &LineMonitor{}
}

// ProcessLine 은 라인 카운터를 증가
func (m *LineMonitor) ProcessLine() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lines++
}

// Update 는 Monitor 인터페이스 구현 (사용하지 않음)
func (m *LineMonitor) Update(bytes int64) {
	// LineMonitor는 ProcessLine을 사용
}

// Total 은 총 처리된 라인 수를 반환
func (m *LineMonitor) Total() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lines
}

// Finish 는 모니터링을 종료
func (m *LineMonitor) Finish() {
	// LineMonitor는 특별한 종료 처리가 필요 없음
}

// Report 는 처리 결과를 문자열로 반환
func (m *LineMonitor) Report() string {
	return fmt.Sprintf("Processed %d lines", m.Total())
}

// TimeMonitor 는 실행 시간을 추적
type TimeMonitor struct {
	startTime time.Time
	endTime   time.Time
	mu        sync.Mutex
}

// NewTimeMonitor 는 새로운 TimeMonitor를 생성
func NewTimeMonitor() *TimeMonitor {
	return &TimeMonitor{}
}

// Start 는 시간 측정을 시작
func (m *TimeMonitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startTime = time.Now()
}

// Update 는 Monitor 인터페이스 구현 (사용하지 않음)
func (m *TimeMonitor) Update(bytes int64) {
	// TimeMonitor는 Start/Finish를 사용
}

// Finish 는 시간 측정을 종료
func (m *TimeMonitor) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.startTime.IsZero() {
		m.endTime = time.Now()
	}
}

// Elapsed 는 경과 시간을 반환
func (m *TimeMonitor) Elapsed() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.startTime.IsZero() {
		return 0
	}
	
	if m.endTime.IsZero() {
		// 아직 진행 중
		return time.Since(m.startTime)
	}
	
	return m.endTime.Sub(m.startTime)
}

// Report 는 실행 시간을 문자열로 반환
func (m *TimeMonitor) Report() string {
	elapsed := m.Elapsed()
	
	if elapsed == 0 {
		return "Time: not started"
	}
	
	if elapsed < time.Second {
		return fmt.Sprintf("Time: %d ms", elapsed.Milliseconds())
	}
	
	return fmt.Sprintf("Time: %.2f seconds", elapsed.Seconds())
}

// humanizeBytes 는 바이트를 읽기 쉬운 형식으로 변환
func humanizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}