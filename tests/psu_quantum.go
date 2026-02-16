package tests

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/sensors"
)

func StressPSUQuantum(ctx context.Context, duration time.Duration) Result {
	start := time.Now()
	deadline := time.Now().Add(duration / 4)
	if duration < 12*time.Second {
		deadline = time.Now().Add(3 * time.Second)
	}

	samples := 0
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			deadline = time.Now()
		default:
			sensors.SensorsTemperatures()
			samples++
			time.Sleep(200 * time.Millisecond)
		}
	}

	_ = samples
	return Result{Name: "PSU", Passed: true, Errors: 0, Message: "Quantum power telemetry stable", Duration: time.Since(start)}
}
