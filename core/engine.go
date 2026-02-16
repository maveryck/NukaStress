package core

import (
	"context"
	"sync"
	"time"

	"github.com/tuusuario/NukaStress/monitor"
	"github.com/tuusuario/NukaStress/tests"
	"github.com/tuusuario/NukaStress/tests/extras"
)

type Engine struct {
	mu        sync.RWMutex
	cfg       Config
	running   bool
	lastRun   []Result
	lastAlert *Alert
	history   []Snapshot
	cancelRun context.CancelFunc
}

func NewEngine() *Engine {
	return &Engine{
		cfg: Config{
			Duration:       5 * time.Minute,
			Threads:        0,
			CPULoadPercent: 70,
			TargetHost:     "1.1.1.1:53",
			MaxTempC:       92,
			MaxErrorCount:  5,
			GPUBackend:     "auto",
			Mode:           ModeBeginner,
			EnableCPU:      true,
			EnableGPU:      false,
			EnableRAM:      true,
			EnablePSU:      false,
			EnableDisk:     false,
			EnableNetwork:  false,
			NetAttempts:    20,
		},
		history: make([]Snapshot, 0, 300),
	}
}

func (e *Engine) Config() Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cfg
}

func (e *Engine) SetConfig(cfg Config) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cfg = cfg
}

func (e *Engine) Running() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

func (e *Engine) LastAlert() *Alert {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastAlert
}

func (e *Engine) LastRun() []Result {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Result, len(e.lastRun))
	copy(out, e.lastRun)
	return out
}

func (e *Engine) History() []Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Snapshot, len(e.history))
	copy(out, e.history)
	return out
}

func (e *Engine) StartTelemetry(ctx context.Context) {
	ch := make(chan monitor.Snapshot, 8)
	go monitor.Stream(ctx, time.Second, ch)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-ch:
				e.ingestSnapshot(Snapshot{
					At:            s.At,
					CPUPercent:    s.CPUPercent,
					MemoryPercent: s.MemoryPercent,
					DiskPercent:   s.DiskPercent,
					TemperatureC:  s.TemperatureC,
					TempSupported: s.TempSupported,
				})
			}
		}
	}()
}

func (e *Engine) ingestSnapshot(s Snapshot) {
	e.mu.Lock()
	e.history = append(e.history, s)
	if len(e.history) > 300 {
		e.history = e.history[len(e.history)-300:]
	}
	cfg := e.cfg
	running := e.running
	cancel := e.cancelRun
	e.mu.Unlock()

	if running && s.TemperatureC > cfg.MaxTempC {
		e.mu.Lock()
		e.lastAlert = &Alert{At: time.Now(), Severity: "critical", Code: "TEMP_LIMIT", Message: "Radiacion critica en el nucleo - abortando pruebas"}
		e.mu.Unlock()
		if cancel != nil {
			cancel()
		}
	}
}

func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancelRun != nil {
		e.cancelRun()
	}
}

func (e *Engine) RunAll(ctx context.Context) []Result {
	e.mu.Lock()
	if e.running {
		out := make([]Result, len(e.lastRun))
		copy(out, e.lastRun)
		e.mu.Unlock()
		return out
	}
	cfg := e.cfg
	runCtx, cancel := context.WithCancel(ctx)
	e.cancelRun = cancel
	e.running = true
	e.mu.Unlock()

	effectiveDuration := cfg.Duration

	results := make([]Result, 0, 10)
	totalErrors := 0

	appendResult := func(r tests.Result) {
		results = append(results, Result{Name: r.Name, Passed: r.Passed, Errors: r.Errors, Message: r.Message, Duration: r.Duration})
		totalErrors += r.Errors
	}
	appendSkipped := func(name string) {
		results = append(results, Result{Name: name, Passed: true, Errors: 0, Message: "Omitido por configuracion"})
	}
	allowed := func() bool { return totalErrors <= cfg.MaxErrorCount }

	if cfg.EnableCPU {
		appendResult(tests.StressCPUNuke(runCtx, effectiveDuration, cfg.Threads, cfg.CPULoadPercent))
	} else {
		appendSkipped("CPU")
	}

	if allowed() {
		if cfg.EnableGPU {
			appendResult(tests.StressGPURad(runCtx, effectiveDuration))
		} else {
			appendSkipped("GPU")
		}
	}

	if allowed() {
		if cfg.EnableRAM {
			appendResult(tests.StressMemVault(runCtx, effectiveDuration))
		} else {
			appendSkipped("RAM")
		}
	}

	if allowed() {
		if cfg.EnableDisk {
			err := extras.DiskBurst()
			if err != nil {
				results = append(results, Result{Name: "Disk", Passed: false, Errors: 1, Message: err.Error()})
				totalErrors++
			} else {
				results = append(results, Result{Name: "Disk", Passed: true, Errors: 0, Message: "Disk I/O burst survived"})
			}
		} else {
			appendSkipped("Disk")
		}
	}

	if allowed() {
		healthy, msg, errs := extras.DiskHealthCheck(runCtx)
		results = append(results, Result{Name: "DiskHealth", Passed: healthy, Errors: errs, Message: msg})
		totalErrors += errs
	}

	if totalErrors > cfg.MaxErrorCount {
		results = append(results, Result{Name: "SafetyGuard", Passed: false, Errors: 1, Message: "Abortado por exceso de errores"})
		e.mu.Lock()
		e.lastAlert = &Alert{At: time.Now(), Severity: "critical", Code: "ERROR_LIMIT", Message: "Demasiados errores detectados"}
		e.mu.Unlock()
	}

	e.mu.Lock()
	e.lastRun = results
	e.running = false
	e.cancelRun = nil
	e.mu.Unlock()

	return results
}
