package core

import "time"

type Mode string

const (
	ModeBeginner  Mode = "beginner"
	ModeWasteland Mode = "wasteland"
)

type Config struct {
	Duration       time.Duration
	Threads        int
	CPULoadPercent int
	TargetHost     string
	MaxTempC       float64
	MaxErrorCount  int
	GPUBackend     string
	Mode           Mode

	EnableCPU     bool
	EnableGPU     bool
	EnableRAM     bool
	EnablePSU     bool
	EnableDisk    bool
	EnableNetwork bool
	NetAttempts   int
}

type Snapshot struct {
	At            time.Time
	CPUPercent    float64
	MemoryPercent float64
	DiskPercent   float64
	TemperatureC  float64
	TempSupported bool
}

type Alert struct {
	At       time.Time
	Severity string
	Code     string
	Message  string
}

type Result struct {
	Name     string
	Passed   bool
	Errors   int
	Message  string
	Duration time.Duration
}
