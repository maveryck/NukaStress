package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tuusuario/NukaStress/core"
)

func WriteJSON(results []core.Result, history []core.Snapshot) (string, error) {
	type payload struct {
		GeneratedAt time.Time       `json:"generated_at"`
		Results     []core.Result   `json:"results"`
		History     []core.Snapshot `json:"history"`
	}

	p := payload{
		GeneratedAt: time.Now(),
		Results:     results,
		History:     history,
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("nukastress_report_%s.json", time.Now().Format("20060102_150405"))
	path := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func WriteCSV(results []core.Result) (string, error) {
	name := fmt.Sprintf("nukastress_report_%s.csv", time.Now().Format("20060102_150405"))
	path := filepath.Join(os.TempDir(), name)

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"test", "status", "errors", "duration", "message"}); err != nil {
		return "", err
	}
	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		record := []string{r.Name, status, fmt.Sprintf("%d", r.Errors), r.Duration.Truncate(time.Millisecond).String(), r.Message}
		if err := w.Write(record); err != nil {
			return "", err
		}
	}

	if err := w.Error(); err != nil {
		return "", err
	}
	return path, nil
}
