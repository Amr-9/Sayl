package tui

import (
	"fmt"
	"time"
)

func fmtDuration(d time.Duration) string {
	if d < time.Millisecond {
		return d.String()
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d)/float64(time.Millisecond))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// Helper for formatting bytes
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatThroughput(bytes int64, durationSeconds float64) string {
	if durationSeconds == 0 {
		return "0 B/s"
	}
	bytesPerSec := float64(bytes) / durationSeconds
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.2f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/1024)
	} else {
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/(1024*1024))
	}
}

func renderSparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}
	levels := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}

	var sb string
	for _, v := range values {
		if max == 0 {
			sb += levels[0]
			continue
		}
		// Calculate index: (v * 7) / max
		idx := (v * 7) / max
		if idx > 7 {
			idx = 7
		}
		sb += levels[idx]
	}
	return sb
}
