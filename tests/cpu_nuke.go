package tests

import (
	"context"
	"math"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

type Result struct {
	Name     string
	Passed   bool
	Errors   int
	Message  string
	Duration time.Duration
}

func StressCPUNuke(ctx context.Context, duration time.Duration, threads int, loadPercent int) Result {
	start := time.Now()
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	if loadPercent <= 0 {
		loadPercent = 10
	}
	if loadPercent > 100 {
		loadPercent = 100
	}
	vendor := detectVendorProfile()

	runCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan int, threads)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(seed float64) {
			defer wg.Done()
			errors := 0
			x := seed
			cycle := 12 * time.Millisecond
			work := time.Duration(loadPercent) * cycle / 100
			if work <= 0 {
				work = 1 * time.Millisecond
			}
			rest := cycle - work
			for {
				select {
				case <-runCtx.Done():
					errCh <- errors
					return
				default:
					deadline := time.Now().Add(work)
					for time.Now().Before(deadline) {
						x = math.Sin(x)*math.Cos(x) + math.Sqrt(math.Abs(x)+1.01)
						if math.IsNaN(x) || math.IsInf(x, 0) {
							errors++
							x = seed + 1
						}
						select {
						case <-runCtx.Done():
							errCh <- errors
							return
						default:
						}
					}
					if rest > 0 {
						select {
						case <-runCtx.Done():
							errCh <- errors
							return
						case <-time.After(rest):
						}
					}
				}
			}
		}(float64(i + 1))
	}

	wg.Wait()
	close(errCh)

	totalErrors := 0
	for e := range errCh {
		totalErrors += e
	}

	msg := "CPU survived the nuclear stress profile"
	if vendor != "unknown" {
		msg = "CPU vendor profile active: " + vendor
	}
	if totalErrors > 0 {
		msg = "CPU instability detected during nuclear run"
	}

	return Result{
		Name:     "CPU",
		Passed:   totalErrors == 0,
		Errors:   totalErrors,
		Message:  msg,
		Duration: time.Since(start),
	}
}

func detectVendorProfile() string {
	info, err := cpu.Info()
	if err != nil || len(info) == 0 {
		return "unknown"
	}
	v := strings.ToLower(info[0].VendorID)
	switch {
	case strings.Contains(v, "intel"):
		return "intel-avx"
	case strings.Contains(v, "amd"):
		return "amd-fma"
	default:
		return "unknown"
	}
}
