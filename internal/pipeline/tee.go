package pipeline

import (
	"io"
	"sync"
)

// TeeWriter 는 입력을 여러 writer에 독립적으로 복사
type TeeWriter struct {
	writers []writerWorker
	wg      sync.WaitGroup
	closed  bool
	mu      sync.Mutex
}

type writerWorker struct {
	writer io.Writer
	queue  chan []byte
}

// NewTeeWriter 는 새로운 TeeWriter를 생성
func NewTeeWriter(writers ...io.Writer) *TeeWriter {
	// Filter out nil writers
	validWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if w != nil {
			validWriters = append(validWriters, w)
		}
	}
	
	t := &TeeWriter{
		writers: make([]writerWorker, len(validWriters)),
	}

	// 각 writer에 독립적인 큐와 고루틴 생성
	for i, w := range validWriters {
		t.writers[i] = writerWorker{
			writer: w,
			queue:  make(chan []byte, 100),
		}
		
		t.wg.Add(1)
		go t.processWriter(i)
	}

	return t
}

// Write 는 데이터를 모든 writer에 비동기적으로 전송
func (t *TeeWriter) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	t.mu.Unlock()

	// 각 writer에게 독립적인 복사본 전송
	for _, w := range t.writers {
		data := make([]byte, len(p))
		copy(data, p)
		
		// Non-blocking 전송
		select {
		case w.queue <- data:
		default:
			// 큐가 가득 차면 블로킹 전송
			w.queue <- data
		}
	}

	return len(p), nil
}

// processWriter 는 각 writer를 독립적으로 처리
func (t *TeeWriter) processWriter(index int) {
	defer t.wg.Done()
	
	w := t.writers[index]
	for data := range w.queue {
		// 에러는 무시 - 다른 writer에 영향을 주지 않음
		w.writer.Write(data)
	}
}

// Close 는 모든 writer가 완료될 때까지 대기
func (t *TeeWriter) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()

	// 모든 큐 닫기
	for _, w := range t.writers {
		close(w.queue)
	}

	// 모든 writer 완료 대기
	t.wg.Wait()

	return nil
}

