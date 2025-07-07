package backup

import (
	"fmt"
	"time"
)

// Progress represents the current state of a backup operation
type Progress struct {
	Current int64  // Current bytes processed
	Total   int64  // Total bytes to process
	Message string // Status message
}

// CalculateSpeed calculates the transfer speed in bytes per second
func (p Progress) CalculateSpeed(duration time.Duration) float64 {
	if duration == 0 {
		return 0
	}
	return float64(p.Current) / duration.Seconds()
}

// CalculateETA calculates the estimated time remaining
func (p Progress) CalculateETA(speed float64) time.Duration {
	if speed == 0 || p.Current >= p.Total {
		return 0
	}
	
	remaining := p.Total - p.Current
	seconds := float64(remaining) / speed
	return time.Duration(seconds * float64(time.Second))
}

// Percentage returns the completion percentage
func (p Progress) Percentage() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Current) / float64(p.Total) * 100
}

// FormatETA formats a duration as a human-readable ETA string
func FormatETA(eta time.Duration) string {
	if eta == 0 {
		return "0s"
	}
	
	hours := int(eta.Hours())
	minutes := int(eta.Minutes()) % 60
	seconds := int(eta.Seconds()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}