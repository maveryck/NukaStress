package extras

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func DiskHealthCheck(ctx context.Context) (bool, string, int) {
	if runtime.GOOS != "windows" {
		return true, "Disk health check no compatible en este SO", 0
	}

	tx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	// Primary path: modern storage cmdlets (may be unavailable on some systems).
	cmdPrimary := exec.CommandContext(tx, "powershell", "-NoProfile", "-Command", "Get-PhysicalDisk | Select-Object FriendlyName,HealthStatus,OperationalStatus | ConvertTo-Csv -NoTypeInformation")
	outPrimary, errPrimary := cmdPrimary.Output()
	if ok, msg, errs, handled := parseDiskHealthCSV(string(outPrimary)); handled {
		return ok, msg, errs
	}
	_ = errPrimary

	// Fallback path: broad compatibility via WMI.
	cmdFallback := exec.CommandContext(tx, "powershell", "-NoProfile", "-Command", "Get-CimInstance -ClassName Win32_DiskDrive | Select-Object Model,Status | ConvertTo-Csv -NoTypeInformation")
	outFallback, errFallback := cmdFallback.Output()
	if ok, msg, errs, handled := parseDiskHealthCSV(string(outFallback)); handled {
		return ok, msg, errs
	}
	if errFallback != nil {
		return true, "Disk SMART/Health no disponible en este equipo", 0
	}

	return true, "Disk SMART/Health sin datos", 0
}

func parseDiskHealthCSV(raw string) (bool, string, int, bool) {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" || !strings.Contains(text, "\"") {
		return true, "", 0, false
	}

	badTokens := []string{"unhealthy", "warning", "lost communication", "failed", "offline", "pred fail", "error"}
	for _, token := range badTokens {
		if strings.Contains(text, token) {
			return false, fmt.Sprintf("Disk SMART/Health reporta estado critico (%s)", token), 1, true
		}
	}

	if strings.Contains(text, "ok") || strings.Contains(text, "healthy") {
		return true, "Disk SMART/Health reportado como saludable", 0, true
	}

	return true, "Disk SMART/Health disponible, sin alertas criticas", 0, true
}
