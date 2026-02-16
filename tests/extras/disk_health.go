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

	tx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	cmd := exec.CommandContext(tx, "powershell", "-NoProfile", "-Command", "Get-PhysicalDisk | Select-Object FriendlyName,HealthStatus,OperationalStatus | ConvertTo-Csv -NoTypeInformation")
	out, err := cmd.Output()
	if err != nil {
		return true, "Disk health no disponible en este equipo", 0
	}

	text := strings.ToLower(strings.TrimSpace(string(out)))
	if text == "" {
		return true, "Disk health sin datos", 0
	}

	badTokens := []string{"unhealthy", "warning", "lost communication", "failed", "offline"}
	for _, token := range badTokens {
		if strings.Contains(text, token) {
			return false, fmt.Sprintf("Disk health reporta estado critico (%s)", token), 1
		}
	}

	return true, "Disk health reportado como saludable", 0
}
