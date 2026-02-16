package tests

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
)

func StressMemVault(ctx context.Context, duration time.Duration) Result {
	start := time.Now()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	bufSize := chooseMemVaultSize()
	buf := make([]byte, bufSize)
	errors := 0
	passes := 0
	deadline := time.Now().Add(duration)

	patterns := []byte{0x00, 0xFF, 0xAA, 0x55}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			deadline = time.Now()
		default:
			for _, p := range patterns {
				for i := range buf {
					buf[i] = p
				}
				for i := 0; i < len(buf); i += 1024 {
					if buf[i] != p {
						errors++
						if errors >= 1024 {
							deadline = time.Now()
							break
						}
					}
				}
				if time.Now().After(deadline) {
					break
				}
			}

			if time.Now().After(deadline) {
				break
			}

			for i := 0; i < 4096; i++ {
				idx := rng.Intn(len(buf))
				v := byte(rng.Intn(256))
				buf[idx] = v
				if buf[idx] != v {
					errors++
				}
				if errors >= 1024 {
					deadline = time.Now()
					break
				}
			}
			passes++
		}
	}

	msg := fmt.Sprintf("RAM advanced pattern check passed (%dMB, %d passes)", bufSize/(1024*1024), passes)
	if errors > 0 {
		msg = fmt.Sprintf("Memory corruption detected (%d errors, %dMB tested)", errors, bufSize/(1024*1024))
	}

	return Result{Name: "RAM", Passed: errors == 0, Errors: errors, Message: msg, Duration: time.Since(start)}
}

func chooseMemVaultSize() int {
	const (
		minSize = 32 * 1024 * 1024
		maxSize = 256 * 1024 * 1024
	)
	size := 64 * 1024 * 1024
	if vm, err := mem.VirtualMemory(); err == nil && vm.Available > 0 {
		target := int(vm.Available / 8)
		if target < minSize {
			target = minSize
		}
		if target > maxSize {
			target = maxSize
		}
		size = target
	}
	return size
}
