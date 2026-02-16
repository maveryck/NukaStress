package gui

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/tuusuario/NukaStress/core"
	"github.com/tuusuario/NukaStress/report"
)

const infiniteProfile = "Infinito (hasta detener)"

func NewMainWindow(a fyne.App) (fyne.Window, error) {
	a.Settings().SetTheme(PipBoyTheme{})

	w := a.NewWindow("NukaStress - Pip-Boy Stress Terminal")
	w.Resize(fyne.NewSize(1220, 780))

	engine := core.NewEngine()
	ctx, cancel := context.WithCancel(context.Background())
	engine.StartTelemetry(ctx)

	var infiniteCancel context.CancelFunc

	status := widget.NewLabel("[STATUS] Idle")
	telemetry := widget.NewLabel("[MONITOR] Pip-Boy online")
	resultsBox := widget.NewMultiLineEntry()
	resultsBox.Disable()
	resultsBox.SetMinRowsVisible(16)

	cpuValue := widget.NewLabel("CPU: 0.0%")
	ramValue := widget.NewLabel("RAM: 0.0%")
	diskValue := widget.NewLabel("DISK: 0.0%")
	tempValue := widget.NewLabel("TEMP: 0.0C")
	tempSupportNote := widget.NewLabel("")
	cpuChart := newMiniChart(100, 180)
	ramChart := newMiniChart(100, 180)
	diskChart := newMiniChart(100, 180)
	tempChart := newMiniChart(120, 180)

	runProfiles := []string{"5 min", "10 min", "15 min", "30 min", "60 min", infiniteProfile}
	runProfile := widget.NewSelect(runProfiles, nil)
	runProfile.SetSelected("10 min")

	exportFormats := []string{"HTML", "JSON", "CSV"}
	selectedExportFormat := "HTML"
	exportFormat := widget.NewSelect(exportFormats, func(v string) {
		if v != "" {
			selectedExportFormat = strings.ToUpper(v)
		}
	})
	exportFormat.SetSelected(selectedExportFormat)

	cfg := engine.Config()
	threadsEntry := widget.NewEntry()
	threadsEntry.SetText(strconv.Itoa(cfg.Threads))
	cpuLoadEntry := widget.NewEntry()
	cpuLoadEntry.SetText(strconv.Itoa(cfg.CPULoadPercent))
	hostEntry := widget.NewEntry()
	hostEntry.SetText(cfg.TargetHost)
	maxTempEntry := widget.NewEntry()
	maxTempEntry.SetText(fmt.Sprintf("%.0f", cfg.MaxTempC))
	maxErrEntry := widget.NewEntry()
	maxErrEntry.SetText(strconv.Itoa(cfg.MaxErrorCount))
	netAttemptsEntry := widget.NewEntry()
	netAttemptsEntry.SetText(strconv.Itoa(cfg.NetAttempts))

	cpuCheck := widget.NewCheck("CPU", nil)
	gpuCheck := widget.NewCheck("GPU", nil)
	ramCheck := widget.NewCheck("RAM", nil)
	psuCheck := widget.NewCheck("PSU", nil)
	diskCheck := widget.NewCheck("Disk", nil)
	netCheck := widget.NewCheck("Network", nil)

	syncUIFromConfig := func(c core.Config) {
		threadsEntry.SetText(strconv.Itoa(c.Threads))
		cpuLoadEntry.SetText(strconv.Itoa(c.CPULoadPercent))
		hostEntry.SetText(c.TargetHost)
		maxTempEntry.SetText(fmt.Sprintf("%.0f", c.MaxTempC))
		maxErrEntry.SetText(strconv.Itoa(c.MaxErrorCount))
		netAttemptsEntry.SetText(strconv.Itoa(c.NetAttempts))
		cpuCheck.SetChecked(c.EnableCPU)
		gpuCheck.SetChecked(c.EnableGPU)
		ramCheck.SetChecked(c.EnableRAM)
		psuCheck.SetChecked(c.EnablePSU)
		diskCheck.SetChecked(c.EnableDisk)
		netCheck.SetChecked(c.EnableNetwork)
	}

	applyConfigFromInputs := func() {
		c := engine.Config()
		if t, err := strconv.Atoi(strings.TrimSpace(threadsEntry.Text)); err == nil {
			c.Threads = t
		}
		if lp, err := strconv.Atoi(strings.TrimSpace(cpuLoadEntry.Text)); err == nil {
			if lp < 10 {
				lp = 10
			}
			if lp > 100 {
				lp = 100
			}
			c.CPULoadPercent = lp
		}
		if host := strings.TrimSpace(hostEntry.Text); host != "" {
			c.TargetHost = host
		}
		if mt, err := strconv.ParseFloat(strings.TrimSpace(maxTempEntry.Text), 64); err == nil && mt > 0 {
			c.MaxTempC = mt
		}
		if me, err := strconv.Atoi(strings.TrimSpace(maxErrEntry.Text)); err == nil && me >= 0 {
			c.MaxErrorCount = me
		}
		if na, err := strconv.Atoi(strings.TrimSpace(netAttemptsEntry.Text)); err == nil && na > 0 {
			c.NetAttempts = na
		}
		c.EnableCPU = cpuCheck.Checked
		c.EnableGPU = gpuCheck.Checked
		c.EnableRAM = ramCheck.Checked
		c.EnablePSU = psuCheck.Checked
		c.EnableDisk = diskCheck.Checked
		c.EnableNetwork = netCheck.Checked
		engine.SetConfig(c)
	}

	cpuCheck.OnChanged = func(bool) { applyConfigFromInputs() }
	gpuCheck.OnChanged = func(bool) { applyConfigFromInputs() }
	ramCheck.OnChanged = func(bool) { applyConfigFromInputs() }
	psuCheck.OnChanged = func(bool) { applyConfigFromInputs() }
	diskCheck.OnChanged = func(bool) { applyConfigFromInputs() }
	netCheck.OnChanged = func(bool) { applyConfigFromInputs() }

	runProfile.OnChanged = func(v string) {
		c := engine.Config()
		if strings.EqualFold(v, infiniteProfile) {
			c.Duration = 0
			engine.SetConfig(c)
			status.SetText("[STATUS] Modo infinito activado")
			return
		}
		d, err := durationFromProfile(v)
		if err != nil {
			status.SetText("[ERROR] Formato de duracion invalido")
			return
		}
		c.Duration = d
		engine.SetConfig(c)
	}

	exportReport := func() {
		results := engine.LastRun()
		if len(results) == 0 {
			status.SetText("[WARN] No hay resultados para exportar")
			return
		}
		history := engine.History()

		var (
			path string
			err  error
		)
		switch selectedExportFormat {
		case "JSON":
			path, err = report.WriteJSON(results, history)
		case "CSV":
			path, err = report.WriteCSV(results)
		default:
			path, err = report.WriteHTML(results, history)
		}
		if err != nil {
			status.SetText("[ERROR] Exportacion fallida: " + err.Error())
			return
		}
		w.Clipboard().SetContent(path)
		status.SetText("[OK] Reporte exportado: " + path + " (copiado)")
	}

	startBtn := widget.NewButton("Iniciar Stress Test", func() {
		if engine.Running() {
			status.SetText("[WARN] Ya hay una corrida activa")
			return
		}
		applyConfigFromInputs()

		if runProfile.Selected == infiniteProfile {
			if infiniteCancel != nil {
				status.SetText("[WARN] El modo infinito ya esta ejecutandose")
				return
			}
			loopCtx, cancelLoop := context.WithCancel(ctx)
			infiniteCancel = cancelLoop
			status.SetText("[STATUS] Modo infinito activo. Usa Detener Test")
			go func() {
				cycle := 1
				for {
					if loopCtx.Err() != nil {
						status.SetText("[STATUS] Modo infinito detenido")
						infiniteCancel = nil
						return
					}
					c := engine.Config()
					if c.Duration <= 0 {
						c.Duration = 10 * time.Minute
						engine.SetConfig(c)
					}
					results := engine.RunAll(loopCtx)
					if len(results) > 0 {
						resultsBox.SetText(formatResultsWithCycle(cycle, results))
						status.SetText(fmt.Sprintf("[STATUS] Ciclo infinito %d finalizado", cycle))
					}
					cycle++
				}
			}()
			return
		}

		if infiniteCancel != nil {
			infiniteCancel()
			infiniteCancel = nil
		}

		status.SetText("[STATUS] Iniciando stress test...")
		go func() {
			results := engine.RunAll(ctx)
			resultsBox.SetText(formatResults(results))
			status.SetText("[STATUS] Stress test finalizado")
		}()
	})

	stopBtn := widget.NewButton("Detener Test", func() {
		if infiniteCancel != nil {
			infiniteCancel()
			infiniteCancel = nil
		}
		engine.Stop()
		status.SetText("[STATUS] Stress test detenido")
	})

	panicBtn := widget.NewButton("Evacuar Vault", func() {
		if infiniteCancel != nil {
			infiniteCancel()
			infiniteCancel = nil
		}
		engine.Stop()
		status.SetText("[CRITICAL] Evacuacion de emergencia activada")
	})

	exportBtn := widget.NewButton("Exportar Reporte", func() { exportReport() })
	applyProBtn := widget.NewButton("Aplicar Pro Config", func() {
		applyConfigFromInputs()
		status.SetText("[OK] Configuracion Pro aplicada")
	})

	startBtn.Importance = widget.HighImportance
	stopBtn.Importance = widget.MediumImportance
	panicBtn.Importance = widget.DangerImportance
	exportBtn.Importance = widget.MediumImportance

	testsPane := container.NewVBox(
		widget.NewCard("Stress Test", "Pruebas de supervivencia para CPU/GPU/RAM/PSU/Disk/Network", container.NewVBox(
			widget.NewLabel("Perfil de duracion"),
			runProfile,
			container.NewGridWithColumns(4, startBtn, stopBtn, panicBtn, exportBtn),
			status,
		)),
		widget.NewCard("Resultados", "Registro del ultimo ciclo", resultsBox),
	)

	cpuPanel := widget.NewCard("CPU", "Uso en vivo", container.NewBorder(cpuValue, nil, nil, nil, cpuChart.Widget()))
	ramPanel := widget.NewCard("RAM", "Uso en vivo", container.NewBorder(ramValue, nil, nil, nil, ramChart.Widget()))
	diskPanel := widget.NewCard("Disk", "Actividad del disco del sistema", container.NewBorder(diskValue, nil, nil, nil, diskChart.Widget()))
	tempPanel := widget.NewCard("Temperatura", "Termica en vivo", container.NewBorder(container.NewVBox(tempValue, tempSupportNote), nil, nil, nil, tempChart.Widget()))

	monitorPane := container.NewVBox(
		widget.NewCard("Monitor", "Vista en tiempo real", telemetry),
		container.NewGridWithColumns(2, cpuPanel, ramPanel),
		container.NewGridWithColumns(2, diskPanel, tempPanel),
	)

	proPane := container.NewVBox(
		widget.NewCard("Pro Config", "Personaliza limites y modulos activos", container.NewVBox(
			widget.NewLabel("Modulos de stress activos"),
			container.NewGridWithColumns(6, cpuCheck, gpuCheck, ramCheck, psuCheck, diskCheck, netCheck),
			widget.NewLabel("Threads CPU (0 = auto)"),
			threadsEntry,
			widget.NewLabel("Carga CPU objetivo (%)"),
			cpuLoadEntry,
			widget.NewLabel("Target host red (host:port)"),
			hostEntry,
			widget.NewLabel("Max temperatura C"),
			maxTempEntry,
			widget.NewLabel("Max errores acumulados"),
			maxErrEntry,
			widget.NewLabel("Intentos de red"),
			netAttemptsEntry,
			applyProBtn,
		)),
	)

	reportsPane := container.NewVBox(
		widget.NewCard("Reports", "Exporta datos del ultimo stress test", container.NewVBox(
			widget.NewLabel("Formato"),
			exportFormat,
			widget.NewButton("Exportar Ahora", exportReport),
			widget.NewLabel("Se copia la ruta al portapapeles."),
		)),
	)

	contentHolder := container.NewMax(testsPane)
	currentPane := "stress"
	showPane := func(o fyne.CanvasObject) {
		contentHolder.Objects = []fyne.CanvasObject{o}
		contentHolder.Refresh()
	}

	navStress := widget.NewButton("", nil)
	navMonitor := widget.NewButton("", nil)
	navPro := widget.NewButton("", nil)
	navReports := widget.NewButton("", nil)

	updateNavState := func() {
		navStress.SetText("[ STRESS TEST ]")
		navMonitor.SetText("[ MONITOR ]")
		navPro.SetText("[ PRO CONFIG ]")
		navReports.SetText("[ REPORTS ]")
		navStress.Importance = widget.MediumImportance
		navMonitor.Importance = widget.MediumImportance
		navPro.Importance = widget.MediumImportance
		navReports.Importance = widget.MediumImportance

		switch currentPane {
		case "monitor":
			navMonitor.SetText(">> MONITOR <<")
			navMonitor.Importance = widget.HighImportance
		case "pro":
			navPro.SetText(">> PRO CONFIG <<")
			navPro.Importance = widget.HighImportance
		case "reports":
			navReports.SetText(">> REPORTS <<")
			navReports.Importance = widget.HighImportance
		default:
			navStress.SetText(">> STRESS TEST <<")
			navStress.Importance = widget.HighImportance
		}
		navStress.Refresh()
		navMonitor.Refresh()
		navPro.Refresh()
		navReports.Refresh()
	}

	navStress.OnTapped = func() {
		currentPane = "stress"
		updateNavState()
		showPane(testsPane)
	}
	navMonitor.OnTapped = func() {
		currentPane = "monitor"
		updateNavState()
		showPane(monitorPane)
	}
	navPro.OnTapped = func() {
		currentPane = "pro"
		updateNavState()
		showPane(proPane)
	}
	navReports.OnTapped = func() {
		currentPane = "reports"
		updateNavState()
		showPane(reportsPane)
	}

	navBar := container.NewGridWithColumns(4, navStress, navMonitor, navPro, navReports)

	footer := widget.NewLabel(fmt.Sprintf("NukaStress // Sobrevive al apocalipsis nuclear de tu hardware // %s/%s", runtime.GOOS, runtime.GOARCH))

	scanOverlay := newScanlineOverlay()
	crtSweep := newCRTSweepOverlay()
	baseLayout := container.NewBorder(navBar, footer, nil, nil, contentHolder)
	displayStack := container.NewStack(baseLayout, scanOverlay, crtSweep.Object())
	w.SetContent(displayStack)

	w.SetCloseIntercept(func() {
		if infiniteCancel != nil {
			infiniteCancel()
		}
		cancel()
		engine.Stop()
		w.Close()
	})

	syncUIFromConfig(engine.Config())
	updateNavState()

	go func() {
		// 100s sweep cycle inspired by CRT-style moving band implementations.
		crtTicker := time.NewTicker(120 * time.Millisecond)
		defer crtTicker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-crtTicker.C:
				crtSweep.Advance(0.0012)
			}
		}
	}()

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				h := engine.History()
				if len(h) > 0 {
					s := h[len(h)-1]
					tempText := "N/A"
					if s.TempSupported {
						tempText = fmt.Sprintf("%.1fC", s.TemperatureC)
						tempSupportNote.SetText("")
					} else {
						tempSupportNote.SetText("Temperatura no compatible en este equipo")
					}
					telemetry.SetText(fmt.Sprintf("[MONITOR] CPU %.1f%% | RAM %.1f%% | DISK %.1f%% | TEMP %s", s.CPUPercent, s.MemoryPercent, s.DiskPercent, tempText))
					cpuValue.SetText(fmt.Sprintf("CPU: %.1f%%", s.CPUPercent))
					ramValue.SetText(fmt.Sprintf("RAM: %.1f%%", s.MemoryPercent))
					diskValue.SetText(fmt.Sprintf("DISK: %.1f%%", s.DiskPercent))
					if s.TempSupported {
						tempValue.SetText(fmt.Sprintf("TEMP: %.1fC", s.TemperatureC))
					} else {
						tempValue.SetText("TEMP: N/A")
					}
					cpuChart.Add(s.CPUPercent)
					ramChart.Add(s.MemoryPercent)
					diskChart.Add(s.DiskPercent)
					if s.TempSupported {
						tempChart.Add(s.TemperatureC)
					}
				}
				if a := engine.LastAlert(); a != nil {
					status.SetText("[" + a.Code + "] " + a.Message)
				}
			}
		}
	}()

	return w, nil
}

func durationFromProfile(profile string) (time.Duration, error) {
	parts := strings.Fields(strings.ToLower(profile))
	if len(parts) == 0 {
		return 0, fmt.Errorf("perfil vacio")
	}
	var mins int
	if _, err := fmt.Sscanf(parts[0], "%d", &mins); err != nil {
		return 0, err
	}
	if mins <= 0 {
		return 0, fmt.Errorf("duracion invalida")
	}
	return time.Duration(mins) * time.Minute, nil
}

func formatResults(results []core.Result) string {
	if len(results) == 0 {
		return "No hay resultados todavia."
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

func formatResultsWithCycle(cycle int, results []core.Result) string {
	return fmt.Sprintf("=== CICLO %d ===\n%s", cycle, formatResults(results))
}
