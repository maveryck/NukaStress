package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/tuusuario/NukaStress/core"
)

func main() {
	minutes := flag.Int("minutes", 10, "duration in minutes")
	flag.Parse()

	engine := core.NewEngine()
	cfg := engine.Config()
	cfg.Duration = time.Duration(*minutes) * time.Minute
	engine.SetConfig(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.StartTelemetry(ctx)

	results := engine.RunAll(ctx)
	fmt.Println("Nukastrest CLI run complete")
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %s | errors=%d | %s\n", status, r.Name, r.Errors, r.Message)
	}
}