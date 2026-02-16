package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/sensors"
)

type Snapshot struct {
	At            time.Time
	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64
	TemperatureC  float64
	TempSupported bool
}

var ioState struct {
	mu        sync.Mutex
	lastAt    time.Time
	lastRead  uint64
	lastWrite uint64
}

func Stream(ctx context.Context, interval time.Duration, out chan<- Snapshot) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			out <- Collect()
		}
	}
}

func Collect() Snapshot {
	s := Snapshot{At: time.Now()}
	if p, err := cpu.Percent(0, false); err == nil && len(p) > 0 {
		s.CPUPercent = p[0]
	}
	if vm, err := mem.VirtualMemory(); err == nil && vm != nil {
		s.MemoryPercent = vm.UsedPercent
	}
	if counters, err := disk.IOCounters(); err == nil && len(counters) > 0 {
		var totalRead uint64
		var totalWrite uint64
		for _, c := range counters {
			totalRead += c.ReadBytes
			totalWrite += c.WriteBytes
		}
		s.DiskPercent = estimateDiskActivity(totalRead, totalWrite, s.At)
	}
	if tmp, err := sensors.SensorsTemperatures(); err == nil && len(tmp) > 0 {
		var sum float64
		var n float64
		for _, v := range tmp {
			if v.Temperature > 0 {
				sum += v.Temperature
				n++
			}
		}
		if n > 0 {
			s.TemperatureC = sum / n
			s.TempSupported = true
		}
	}
	return s
}

func estimateDiskActivity(totalRead, totalWrite uint64, at time.Time) float64 {
	ioState.mu.Lock()
	defer ioState.mu.Unlock()

	if ioState.lastAt.IsZero() {
		ioState.lastAt = at
		ioState.lastRead = totalRead
		ioState.lastWrite = totalWrite
		return 0
	}

	elapsed := at.Sub(ioState.lastAt).Seconds()
	if elapsed <= 0 {
		return 0
	}
	var deltaRead uint64
	if totalRead >= ioState.lastRead {
		deltaRead = totalRead - ioState.lastRead
	}
	var deltaWrite uint64
	if totalWrite >= ioState.lastWrite {
		deltaWrite = totalWrite - ioState.lastWrite
	}
	ioState.lastAt = at
	ioState.lastRead = totalRead
	ioState.lastWrite = totalWrite

	bytesPerSec := float64(deltaRead+deltaWrite) / elapsed
	// 300 MB/s baseline for "100%" activity indicator (throughput-based).
	activity := (bytesPerSec / (300 * 1024 * 1024)) * 100.0
	if activity < 0 {
		return 0
	}
	if activity > 100 {
		return 100
	}
	return activity
}
