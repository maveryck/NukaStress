package gui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type miniChart struct {
	max      float64
	values   []float64
	capacity int
	raster   *canvas.Raster
}

func newMiniChart(max float64, capacity int) *miniChart {
	if max <= 0 {
		max = 100
	}
	if capacity <= 0 {
		capacity = 120
	}
	m := &miniChart{max: max, capacity: capacity, values: make([]float64, 0, capacity)}
	m.raster = canvas.NewRaster(func(w, h int) image.Image {
		return m.render(w, h)
	})
	m.raster.SetMinSize(fyne.NewSize(420, 160))
	return m
}

func (m *miniChart) Add(v float64) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	if v < 0 {
		v = 0
	}
	if v > m.max {
		v = m.max
	}
	m.values = append(m.values, v)
	if len(m.values) > m.capacity {
		m.values = m.values[len(m.values)-m.capacity:]
	}
	m.raster.Refresh()
}

func (m *miniChart) Widget() *canvas.Raster {
	return m.raster
}

func (m *miniChart) render(w, h int) image.Image {
	if w <= 0 {
		w = 420
	}
	if h <= 0 {
		h = 160
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	bg := color.NRGBA{R: 8, G: 16, B: 8, A: 220}
	grid := color.NRGBA{R: 34, G: 70, B: 34, A: 160}
	line := color.NRGBA{R: 120, G: 255, B: 120, A: 255}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, bg)
		}
	}

	for i := 1; i <= 4; i++ {
		y := (h - 1) * i / 5
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, grid)
		}
	}

	for i := 1; i <= 5; i++ {
		x := (w - 1) * i / 6
		for y := 0; y < h; y++ {
			img.SetNRGBA(x, y, grid)
		}
	}

	if len(m.values) < 2 {
		return img
	}

	prevX := 0
	prevY := valueToY(m.values[0], m.max, h)
	for x := 1; x < w; x++ {
		idx := int(float64(x) / float64(w-1) * float64(len(m.values)-1))
		y := valueToY(m.values[idx], m.max, h)
		drawLine(img, prevX, prevY, x, y, line)
		prevX, prevY = x, y
	}

	return img
}

func valueToY(v, max float64, h int) int {
	n := v / max
	if n < 0 {
		n = 0
	}
	if n > 1 {
		n = 1
	}
	y := int((1 - n) * float64(h-1))
	if y < 0 {
		y = 0
	}
	if y >= h {
		y = h - 1
	}
	return y
}

func drawLine(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	dx := int(math.Abs(float64(x1 - x0)))
	dy := -int(math.Abs(float64(y1 - y0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy

	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.SetNRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}
