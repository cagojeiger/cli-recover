package pipeline

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// SlowWriter 는 느린 writer를 시뮬레이션
type SlowWriter struct {
	buf   bytes.Buffer
	delay time.Duration
	mu    sync.Mutex
}

func (w *SlowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(w.delay)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *SlowWriter) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Bytes()
}

func (w *SlowWriter) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Len()
}

func TestTeeWriter(t *testing.T) {
	t.Run("writes to single writer", func(t *testing.T) {
		var buf bytes.Buffer
		tee := NewTeeWriter(&buf)

		data := []byte("test data")
		n, err := tee.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		
		// 비동기 처리를 위한 짧은 대기
		time.Sleep(10 * time.Millisecond)
		
		assert.Equal(t, data, buf.Bytes())
	})

	t.Run("writes to multiple writers", func(t *testing.T) {
		var buf1, buf2, buf3 bytes.Buffer
		tee := NewTeeWriter(&buf1, &buf2, &buf3)

		data := []byte("multiple writers test")
		n, err := tee.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		
		// 비동기 처리를 위한 짧은 대기
		time.Sleep(10 * time.Millisecond)
		
		assert.Equal(t, data, buf1.Bytes())
		assert.Equal(t, data, buf2.Bytes())
		assert.Equal(t, data, buf3.Bytes())
	})

	t.Run("ensures writer independence", func(t *testing.T) {
		// 느린 writer와 빠른 writer 준비
		slowWriter := &SlowWriter{delay: 100 * time.Millisecond}
		fastWriter := &bytes.Buffer{}

		tee := NewTeeWriter(slowWriter, fastWriter)

		data := []byte("independence test")
		n, err := tee.Write(data)

		// Write는 즉시 반환해야 함
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		// 빠른 writer는 즉시 데이터를 가져야 함
		// (독립적 실행이므로 약간의 지연 허용)
		time.Sleep(10 * time.Millisecond)
		assert.Equal(t, data, fastWriter.Bytes())

		// 느린 writer는 아직 데이터가 없을 수 있음
		assert.Equal(t, 0, slowWriter.Len())

		// 충분히 대기 후 느린 writer도 데이터를 가져야 함
		time.Sleep(150 * time.Millisecond)
		assert.Equal(t, data, slowWriter.Bytes())
	})

	t.Run("handles multiple writes", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer
		tee := NewTeeWriter(&buf1, &buf2)

		// 여러 번 쓰기
		data1 := []byte("first ")
		data2 := []byte("second ")
		data3 := []byte("third")

		tee.Write(data1)
		tee.Write(data2)
		tee.Write(data3)

		expected := []byte("first second third")
		
		// 약간의 지연 후 확인 (비동기 처리)
		time.Sleep(50 * time.Millisecond)
		
		assert.Equal(t, expected, buf1.Bytes())
		assert.Equal(t, expected, buf2.Bytes())
	})

	t.Run("close waits for all writers", func(t *testing.T) {
		slowWriter1 := &SlowWriter{delay: 50 * time.Millisecond}
		slowWriter2 := &SlowWriter{delay: 100 * time.Millisecond}

		tee := NewTeeWriter(slowWriter1, slowWriter2)

		data := []byte("close test")
		tee.Write(data)

		// Close는 모든 writer가 완료될 때까지 대기해야 함
		start := time.Now()
		err := tee.Close()
		elapsed := time.Since(start)

		assert.NoError(t, err)
		// 최소 100ms는 걸려야 함 (가장 느린 writer)
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(100))

		// 모든 writer가 데이터를 받았는지 확인
		assert.Equal(t, data, slowWriter1.Bytes())
		assert.Equal(t, data, slowWriter2.Bytes())
	})

	t.Run("handles writer errors gracefully", func(t *testing.T) {
		// 에러를 반환하는 writer
		errorWriter := &ErrorWriter{err: io.ErrShortWrite}
		normalWriter := &bytes.Buffer{}

		tee := NewTeeWriter(errorWriter, normalWriter)

		data := []byte("error handling test")
		n, err := tee.Write(data)

		// Write 자체는 성공해야 함 (독립적 실행)
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)

		// 정상 writer는 데이터를 받아야 함
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, data, normalWriter.Bytes())
	})

	t.Run("concurrent writes are safe", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer
		tee := NewTeeWriter(&buf1, &buf2)

		var wg sync.WaitGroup
		numGoroutines := 10
		dataPerGoroutine := 100

		// 동시에 여러 고루틴에서 쓰기
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < dataPerGoroutine; j++ {
					data := []byte{byte(id)}
					tee.Write(data)
				}
			}(i)
		}

		wg.Wait()
		tee.Close()

		// 각 writer가 모든 데이터를 받았는지 확인
		assert.Equal(t, numGoroutines*dataPerGoroutine, buf1.Len())
		assert.Equal(t, numGoroutines*dataPerGoroutine, buf2.Len())
	})
}

// ErrorWriter 는 항상 에러를 반환하는 writer
type ErrorWriter struct {
	err error
}

func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestTeeWriter_ErrorHandling(t *testing.T) {
	t.Run("write after close returns error", func(t *testing.T) {
		var buf bytes.Buffer
		tee := NewTeeWriter(&buf)
		
		// Close the writer
		err := tee.Close()
		assert.NoError(t, err)
		
		// Try to write after close
		data := []byte("should not write")
		n, err := tee.Write(data)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "closed pipe")
		assert.Equal(t, 0, n)
		
		// Buffer should be empty
		assert.Equal(t, 0, buf.Len())
	})
	
	t.Run("close multiple times", func(t *testing.T) {
		var buf bytes.Buffer
		tee := NewTeeWriter(&buf)
		
		// Write some data
		tee.Write([]byte("test"))
		
		// First close
		err := tee.Close()
		assert.NoError(t, err)
		
		// Second close should be safe
		err = tee.Close()
		assert.NoError(t, err)
	})
	
	t.Run("handles nil writer", func(t *testing.T) {
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewTeeWriter panicked with nil writer: %v", r)
			}
		}()
		
		// Create with nil writer should be handled gracefully
		var buf bytes.Buffer
		tee := NewTeeWriter(nil, &buf)
		
		data := []byte("test with nil")
		n, err := tee.Write(data)
		
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		
		// Wait for async processing
		time.Sleep(10 * time.Millisecond)
		
		// Only the valid writer should have data
		assert.Equal(t, data, buf.Bytes())
		
		// Close should also work
		err = tee.Close()
		assert.NoError(t, err)
	})
	
	t.Run("queue overflow handling", func(t *testing.T) {
		// Use a very slow writer that will cause queue to back up
		slowWriter := &SlowWriter{delay: 100 * time.Millisecond}
		tee := NewTeeWriter(slowWriter)
		
		// Write a lot of data quickly
		for i := 0; i < 200; i++ {
			data := []byte("overflow test data")
			n, err := tee.Write(data)
			assert.NoError(t, err)
			assert.Equal(t, len(data), n)
		}
		
		// Close should wait for all data to be processed
		err := tee.Close()
		assert.NoError(t, err)
		
		// All data should have been written
		expectedLen := 200 * len("overflow test data")
		assert.Equal(t, expectedLen, slowWriter.Len())
	})
}

func TestTeeWriter_RealWorldScenarios(t *testing.T) {
	t.Run("pipeline with monitoring", func(t *testing.T) {
		// Simulate a real pipeline scenario
		var output bytes.Buffer
		var logFile bytes.Buffer
		monitor := NewByteMonitor()
		monitorWriter := NewMonitorWriter(monitor)
		
		tee := NewTeeWriter(&output, &logFile, monitorWriter)
		
		// Simulate pipeline data
		data := []byte("Pipeline output data\nLine 2\nLine 3\n")
		n, err := tee.Write(data)
		
		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		
		// Close and wait
		err = tee.Close()
		assert.NoError(t, err)
		
		// All writers should have the data
		assert.Equal(t, data, output.Bytes())
		assert.Equal(t, data, logFile.Bytes())
		assert.Equal(t, int64(len(data)), monitor.Total())
	})
	
	t.Run("checksum calculation", func(t *testing.T) {
		// Simulate checksum calculation during write
		var mainOutput bytes.Buffer
		checksumWriter := NewChecksumWriter("sha256")
		
		tee := NewTeeWriter(&mainOutput, checksumWriter)
		
		// Write data in chunks
		chunks := [][]byte{
			[]byte("chunk1 "),
			[]byte("chunk2 "),
			[]byte("chunk3"),
		}
		
		for _, chunk := range chunks {
			n, err := tee.Write(chunk)
			assert.NoError(t, err)
			assert.Equal(t, len(chunk), n)
		}
		
		err := tee.Close()
		assert.NoError(t, err)
		
		// Verify both outputs
		expected := []byte("chunk1 chunk2 chunk3")
		assert.Equal(t, expected, mainOutput.Bytes())
		
		// Checksum should be calculated correctly
		checksum := checksumWriter.Sum()
		assert.NotEmpty(t, checksum)
	})
}