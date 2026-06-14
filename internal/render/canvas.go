// Package render builds the wallpaper image by alpha-compositing layers onto a
// single equirectangular master canvas, then cropping to the display. It uses
// only the standard library image/draw plus golang.org/x/image/draw, so there
// is no heavy plotting/geo stack.
package render

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/ddnewell/geis-wallpaper/internal/geo"
)

// Canvas is the master RGBA image plus its geographic grid.
type Canvas struct {
	Img  *image.RGBA
	Grid geo.Grid
}

// NewCanvas allocates a w x h master canvas.
func NewCanvas(w, h int) *Canvas {
	return &Canvas{
		Img:  image.NewRGBA(image.Rect(0, 0, w, h)),
		Grid: geo.Grid{W: w, H: h},
	}
}

// DrawOver composites src fully opaque over the canvas at the origin.
func (c *Canvas) DrawOver(src image.Image) {
	draw.Draw(c.Img, c.Img.Bounds(), src, src.Bounds().Min, draw.Over)
}

// DrawAlpha composites src over the canvas using a uniform global alpha (0..1).
func (c *Canvas) DrawAlpha(src image.Image, alpha float64) {
	a := uint8(clamp01(alpha) * 255)
	mask := image.NewUniform(color.Alpha{A: a})
	draw.DrawMask(c.Img, c.Img.Bounds(), src, src.Bounds().Min, mask, image.Point{}, draw.Over)
}

// FillRect composites a translucent rectangle (NRGBA) over the canvas.
func (c *Canvas) FillRect(r image.Rectangle, col color.NRGBA) {
	draw.Draw(c.Img, r, image.NewUniform(col), image.Point{}, draw.Over)
}

// DrawDisc composites a filled, anti-aliased-ish disc of radius r centered at
// (cx,cy) using non-premultiplied NRGBA color (with its own alpha).
func DrawDisc(dst *image.RGBA, cx, cy, r int, col color.NRGBA) {
	if r < 1 {
		r = 1
	}
	tile := image.NewNRGBA(image.Rect(0, 0, 2*r+1, 2*r+1))
	r2 := float64(r * r)
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			d2 := float64(dx*dx + dy*dy)
			if d2 <= r2 {
				tile.SetNRGBA(dx+r, dy+r, col)
			}
		}
	}
	draw.Draw(dst, image.Rect(cx-r, cy-r, cx+r+1, cy+r+1), tile, image.Point{}, draw.Over)
}

// DrawDiamond composites a filled diamond (rotated square) marker.
func DrawDiamond(dst *image.RGBA, cx, cy, r int, col color.NRGBA) {
	if r < 1 {
		r = 1
	}
	tile := image.NewNRGBA(image.Rect(0, 0, 2*r+1, 2*r+1))
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if abs(dx)+abs(dy) <= r {
				tile.SetNRGBA(dx+r, dy+r, col)
			}
		}
	}
	draw.Draw(dst, image.Rect(cx-r, cy-r, cx+r+1, cy+r+1), tile, image.Point{}, draw.Over)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
