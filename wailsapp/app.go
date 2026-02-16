package wailsapp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tuusuario/NukaStress/core"
	"github.com/tuusuario/NukaStress/report"
)

type SnapshotDTO struct {
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent"`
	DiskPercent   float64 `json:"diskPercent"`
	TemperatureC  float64 `json:"temperatureC"`
	TempSupported bool    `json:"tempSupported"`
}

type StatusDTO struct {
	Running     bool        `json:"running"`
	Mode        string      `json:"mode"`
	StatusText  string      `json:"statusText"`
	Snapshot    SnapshotDTO `json:"snapshot"`
	LastResults string      `json:"lastResults"`
	Findings    []string    `json:"findings"`
}

type App struct {
	ctx         context.Context
	engine      *core.Engine
	telemetryCx context.CancelFunc
	mu          sync.Mutex
}

func New() *App {
	return &App{engine: core.NewEngine()}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	tx, cancel := context.WithCancel(context.Background())
	a.telemetryCx = cancel
	a.engine.StartTelemetry(tx)
}

func (a *App) Shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.telemetryCx != nil {
		a.telemetryCx()
		a.telemetryCx = nil
	}
	a.engine.Stop()
}

func (a *App) SetConfig(cpu, gpu, ram, disk, psu, network bool, cpuLoadPercent int, minutes int) {
	cfg := a.engine.Config()
	cfg.EnableCPU = cpu
	cfg.EnableGPU = gpu
	cfg.EnableRAM = ram
	cfg.EnableDisk = disk
	cfg.EnablePSU = false
	cfg.EnableNetwork = false
	_ = psu
	_ = network
	if cpuLoadPercent < 10 {
		cpuLoadPercent = 10
	}
	if cpuLoadPercent > 100 {
		cpuLoadPercent = 100
	}
	cfg.CPULoadPercent = cpuLoadPercent
	if minutes > 0 {
		cfg.Duration = time.Duration(minutes) * time.Minute
	}
	a.engine.SetConfig(cfg)
}

func (a *App) StartStress() string {
	if a.engine.Running() {
		return "Ya hay una corrida en progreso"
	}
	go func() {
		a.engine.RunAll(context.Background())
	}()
	return "Stress test iniciado"
}

func (a *App) StopStress() string {
	a.engine.Stop()
	return "Stress test detenido"
}

func (a *App) ExportReport(format string) (string, error) {
	results := a.engine.LastRun()
	if len(results) == 0 {
		return "", fmt.Errorf("no hay resultados para exportar")
	}
	history := a.engine.History()
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return report.WriteJSON(results, history)
	case "csv":
		return report.WriteCSV(results)
	default:
		return report.WriteHTML(results, history)
	}
}

func (a *App) GetStatus() StatusDTO {
	cfg := a.engine.Config()
	history := a.engine.History()
	run := a.engine.LastRun()

	s := SnapshotDTO{}
	if len(history) > 0 {
		last := history[len(history)-1]
		s = SnapshotDTO{
			CPUPercent:    last.CPUPercent,
			MemoryPercent: last.MemoryPercent,
			DiskPercent:   last.DiskPercent,
			TemperatureC:  last.TemperatureC,
			TempSupported: last.TempSupported,
		}
	}

	statusText := "Idle"
	if a.engine.Running() {
		statusText = "Running"
	}
	if alert := a.engine.LastAlert(); alert != nil {
		statusText = "[" + alert.Code + "] " + alert.Message
	}

	return StatusDTO{
		Running:     a.engine.Running(),
		Mode:        string(cfg.Mode),
		StatusText:  statusText,
		Snapshot:    s,
		LastResults: formatResults(run),
		Findings:    collectFindings(run),
	}
}

func collectFindings(results []core.Result) []string {
	findings := make([]string, 0, len(results))
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.Message), "omitido") {
			continue
		}
		if r.Errors > 0 || !r.Passed {
			findings = append(findings, fmt.Sprintf("%s: %s", r.Name, r.Message))
		}
	}
	if len(findings) == 0 && len(results) > 0 {
		findings = append(findings, "Sin fallos detectados en la ultima corrida")
	}
	return findings
}

func formatResults(results []core.Result) string {
	if len(results) == 0 {
		return "No hay resultados todavia"
	}
	var b strings.Builder
	for _, r := range results {
		st := "PASS"
		if !r.Passed {
			st = "FAIL"
		}
		b.WriteString(fmt.Sprintf("[%s] %s | errors=%d | %s\n", st, r.Name, r.Errors, r.Message))
	}
	return b.String()
}
