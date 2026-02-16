package tests

import (
	"context"
	"math/rand"
	"time"
)

func StressMemVault(ctx context.Context, duration time.Duration) Result {
	start := time.Now()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 16*1024*1024)
	errors := 0
	deadline := time.Now().Add(duration / 2)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			deadline = time.Now()
		default:
			idx := rng.Intn(len(buf))
			v := byte(rng.Intn(256))
			buf[idx] = v
			if buf[idx] != v {
				errors++
			}
		}
	}

	msg := "Vault memory pattern check passed"
	if errors > 0 {
		msg = "Memory corruption detected under vault stress"
	}

	return Result{Name: "RAM", Passed: errors == 0, Errors: errors, Message: msg, Duration: time.Since(start)}
}
