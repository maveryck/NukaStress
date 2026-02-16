package gui

import (
	"fmt"
	"math"

	"github.com/tuusuario/NukaStress/core"
)

func makeSparkline(history []core.Snapshot, field func(core.Snapshot) float64, width int) string {
	if width <= 0 {
		width = 40
	}
	if len(history) == 0 {
		return "(sin datos)"
	}
	if len(history) > width {
		history = history[len(history)-width:]
	}

	chars := []rune(" .:-=+*#%@")
	minV := math.MaxFloat64
	maxV := -math.MaxFloat64
	values := make([]float64, 0, len(history))

	for _, s := range history {
		v := field(s)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = 0
		}
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
		values = append(values, v)
	}

	if maxV-minV < 0.0001 {
		return fmt.Sprintf("%s", string(chars[len(chars)/2]))
	}

	out := make([]rune, 0, len(values))
	for _, v := range values {
		norm := (v - minV) / (maxV - minV)
		idx := int(norm * float64(len(chars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		out = append(out, chars[idx])
	}
	return string(out)
}
