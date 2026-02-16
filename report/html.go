package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tuusuario/NukaStress/core"
)

func WriteHTML(results []core.Result, history []core.Snapshot) (string, error) {
	name := fmt.Sprintf("nukastress_report_%s.html", time.Now().Format("20060102_150405"))
	path := filepath.Join(os.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b := &strings.Builder{}
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>NukaStress Report</title>")
	b.WriteString("<style>body{font-family:Segoe UI,Arial,sans-serif;background:#0b0f0b;color:#b8ffbf;padding:24px}table{width:100%;border-collapse:collapse}th,td{border:1px solid #245a2a;padding:8px}th{background:#143017}.ok{color:#72ff8a}.bad{color:#ff4b4b}.finding{margin:.3rem 0;color:#ffb4b4}</style>")
	b.WriteString("</head><body><h1>NukaStress - Rad Report</h1>")
	b.WriteString("<p>Test Nuclear completado. Sobrevivientes en el resumen inferior.</p>")
	b.WriteString("<h2>Fallos Detectados</h2><ul>")
	for _, line := range collectFindings(results) {
		b.WriteString(fmt.Sprintf("<li class=\"finding\">%s</li>", line))
	}
	b.WriteString("</ul>")
	b.WriteString("<table><thead><tr><th>Test</th><th>Status</th><th>Errors</th><th>Duration</th><th>Message</th></tr></thead><tbody>")
	for _, r := range results {
		st := "PASS"
		cls := "ok"
		if !r.Passed {
			st = "FAIL"
			cls = "bad"
		}
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td class=\"%s\">%s</td><td>%d</td><td>%s</td><td>%s</td></tr>", r.Name, cls, st, r.Errors, r.Duration.Truncate(time.Millisecond), r.Message))
	}
	b.WriteString("</tbody></table>")
	if len(history) > 0 {
		last := history[len(history)-1]
		tempText := "N/A"
		if last.TempSupported {
			tempText = fmt.Sprintf("%.1fC", last.TemperatureC)
		}
		b.WriteString(fmt.Sprintf("<p>Ultimo estado: CPU %.1f%% | RAM %.1f%% | DISK %.1f%% | TEMP %s</p>", last.CPUPercent, last.MemoryPercent, last.DiskPercent, tempText))
	}
	b.WriteString("</body></html>")

	if _, err := f.WriteString(b.String()); err != nil {
		return "", err
	}
	return path, nil
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
	if len(findings) == 0 {
		return []string{"Sin fallos detectados en la ultima corrida"}
	}
	return findings
}
