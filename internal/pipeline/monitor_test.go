package pipeline

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestByteMonitor(t *testing.T) {
	t.Run("tracks total bytes", func(t *testing.T) {
		monitor := NewByteMonitor()

		// 100 바이트 처리
		monitor.Update(100)
		assert.Equal(t, int64(100), monitor.Total())

		// 추가 50 바이트
		monitor.Update(50)
		assert.Equal(t, int64(150), monitor.Total())
	})

	t.Run("thread safe updates", func(t *testing.T) {
		monitor := NewByteMonitor()
		var wg sync.WaitGroup

		// 10개의 고루틴에서 동시에 업데이트
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					monitor.Update(1)
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(1000), monitor.Total())
	})

	t.Run("generates report", func(t *testing.T) {
		monitor := NewByteMonitor()
		monitor.Update(1024) // 1KB

		report := monitor.Report()
		assert.Contains(t, report, "1024 bytes")
		assert.Contains(t, report, "(1.0 KB)")
	})

	t.Run("finish stops monitoring", func(t *testing.T) {
		monitor := NewByteMonitor()
		monitor.Update(100)
		monitor.Finish()

		// Finish 후에도 Total은 유지됨
		assert.Equal(t, int64(100), monitor.Total())
	})
}

func TestLineMonitor(t *testing.T) {
	t.Run("counts lines", func(t *testing.T) {
		monitor := NewLineMonitor()

		// Process 호출마다 라인 증가
		monitor.ProcessLine()
		assert.Equal(t, int64(1), monitor.Total())

		monitor.ProcessLine()
		monitor.ProcessLine()
		assert.Equal(t, int64(3), monitor.Total())
	})

	t.Run("thread safe line counting", func(t *testing.T) {
		monitor := NewLineMonitor()
		var wg sync.WaitGroup

		// 동시에 라인 처리
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					monitor.ProcessLine()
				}
			}()
		}

		wg.Wait()
		assert.Equal(t, int64(100), monitor.Total())
	})

	t.Run("generates line report", func(t *testing.T) {
		monitor := NewLineMonitor()
		for i := 0; i < 42; i++ {
			monitor.ProcessLine()
		}

		report := monitor.Report()
		assert.Contains(t, report, "42 lines")
	})
}

func TestTimeMonitor(t *testing.T) {
	t.Run("tracks elapsed time", func(t *testing.T) {
		monitor := NewTimeMonitor()
		monitor.Start()

		// 짧은 대기
		time.Sleep(100 * time.Millisecond)

		monitor.Finish()
		elapsed := monitor.Elapsed()

		// 최소 100ms 이상이어야 함
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(100))
		// 하지만 200ms 미만이어야 함 (여유 마진)
		assert.Less(t, elapsed.Milliseconds(), int64(200))
	})

	t.Run("generates time report", func(t *testing.T) {
		monitor := NewTimeMonitor()
		monitor.Start()
		time.Sleep(50 * time.Millisecond)
		monitor.Finish()

		report := monitor.Report()
		assert.Contains(t, report, "ms")
	})

	t.Run("handles not started", func(t *testing.T) {
		monitor := NewTimeMonitor()
		
		// Start 없이 Elapsed 호출
		elapsed := monitor.Elapsed()
		assert.Equal(t, time.Duration(0), elapsed)

		report := monitor.Report()
		assert.Contains(t, report, "not started")
	})
}