package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2/app"

	"github.com/tuusuario/NukaStress/core"
	"github.com/tuusuario/NukaStress/gui"
)

func main() {
	headless := flag.Bool("headless", false, "run without GUI")
	minutes := flag.Int("minutes", 10, "duration for headless run")
	flag.Parse()

	if *headless {
		runHeadless(*minutes)
		return
	}

	a := app.NewWithID("com.nukastress.app")
	w, err := gui.NewMainWindow(a)
	if err != nil {
		log.Fatalf("failed to initialize NukaStress UI: %v", err)
	}
	w.ShowAndRun()
}

func runHeadless(minutes int) {
	engine := core.NewEngine()
	cfg := engine.Config()
	cfg.Duration = time.Duration(minutes) * time.Minute
	engine.SetConfig(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.StartTelemetry(ctx)

	results := engine.RunAll(ctx)
	fmt.Println("NukaStress headless run complete")
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %s | errors=%d | %s\n", status, r.Name, r.Errors, r.Message)
	}
}
