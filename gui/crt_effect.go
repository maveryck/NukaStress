package gui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

// newScanlineOverlay emulates a static scanline mask + subtle edge glow.
func newScanlineOverlay() *canvas.Raster {
	return canvas.NewRaster(func(w, h int) image.Image {
		if w <= 0 {
			w = 1
		}
		if h <= 0 {
			h = 1
		}

		img := image.NewNRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			lineAlpha := uint8(0)
			// 2px pattern: one darker line, one clear line.
			if y%2 == 0 {
				lineAlpha = 12
			}

			// Vertical vignette similar to CRT edge backlight falloff.
			edge := math.Abs((float64(y)/float64(h))*2.0 - 1.0)
			vignette := uint8(0)
			if edge > 0.78 {
				v := int((edge - 0.78) * 120)
				if v > 28 {
					v = 28
				}
				vignette = uint8(v)
			}

			a := lineAlpha + vignette
			if a == 0 {
				continue
			}
			for x := 0; x < w; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 12, G: 255, B: 90, A: a})
			}
		}
		return img
	})
}

type crtSweep struct {
	phase  float64
	raster *canvas.Raster
}

func newCRTSweepOverlay() *crtSweep {
	c := &crtSweep{}
	c.raster = canvas.NewRaster(func(w, h int) image.Image {
		if w <= 0 {
			w = 1
		}
		if h <= 0 {
			h = 1
		}
		img := image.NewNRGBA(image.Rect(0, 0, w, h))

		centerY := int(c.phase * float64(h))
		top := centerY - 28
		bottom := centerY + 28
		if top < 0 {
			top = 0
		}
		if bottom >= h {
			bottom = h - 1
		}

		for y := top; y <= bottom; y++ {
			d := math.Abs(float64(y - centerY))
			alpha := int((28 - d) * 1.2)
			if alpha < 0 {
				alpha = 0
			}
			if alpha > 24 {
				alpha = 24
			}
			if alpha == 0 {
				continue
			}
			for x := 0; x < w; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 190, G: 255, B: 190, A: uint8(alpha)})
			}
		}
		return img
	})
	return c
}

func (c *crtSweep) Object() *canvas.Raster {
	return c.raster
}

func (c *crtSweep) Advance(delta float64) {
	c.phase = math.Mod(c.phase+delta, 1.0)
	c.raster.Refresh()
}
