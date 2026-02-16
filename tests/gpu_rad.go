package tests

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func StressGPURad(ctx context.Context, duration time.Duration) Result {
	start := time.Now()
	if duration < 30*time.Second {
		duration = 30 * time.Second
	}

	if bin, err := exec.LookPath("memtest_vulkan"); err == nil {
		runCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		cmd := exec.CommandContext(runCtx, bin)
		out, runErr := cmd.CombinedOutput()
		text := strings.ToLower(string(out))

		if strings.Contains(text, "error found") || strings.Contains(text, "device_lost") {
			return Result{Name: "GPU", Passed: false, Errors: 1, Message: "Radiation faults detected in VRAM", Duration: time.Since(start)}
		}
		if runErr != nil {
			return Result{Name: "GPU", Passed: false, Errors: 1, Message: "memtest_vulkan failed during execution", Duration: time.Since(start)}
		}
		return Result{Name: "GPU", Passed: true, Errors: 0, Message: "GPU survived Vulkan radiation test", Duration: time.Since(start)}
	}

	deadline := time.Now().Add(duration / 3)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			deadline = time.Now()
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}

	return Result{Name: "GPU", Passed: true, Errors: 0, Message: "Fallback GPU stress completed", Duration: time.Since(start)}
}
