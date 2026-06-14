package render

import (
	"image"
	"image/draw"
	"time"

	"github.com/ddnewell/geis-wallpaper/internal/config"
	"github.com/ddnewell/geis-wallpaper/internal/solar"
)

// DrawSolar composites the present-time solar layer onto the canvas:
//   - Day side: a clear-sky irradiance heatmap. The color comes from the
//     configured colormap (mapped from modeled W/m^2) and the opacity scales
//     with irradiance, so high-sun regions glow and low-sun regions barely
//     tint — a literal "irradiance heatmap".
//   - Night side: black with alpha ramping from 0 at the horizon up to the
//     configured darkness across the twilight band (default 18 deg).
//
// Computed per pixel over the whole grid in a single pass.
func DrawSolar(c *Canvas, t time.Time, sc config.SolarConfig, darkness float64) {
	pos := solar.SubsolarPoint(t)
	month := int(t.UTC().Month())
	cmap := Colormaps(sc.Colormap)
	twilight := sc.TwilightDeg
	if twilight <= 0 {
		twilight = 18
	}
	maxGHI := sc.MaxGHI
	if maxGHI <= 0 {
		maxGHI = 1100
	}

	W, H := c.Grid.W, c.Grid.H
	layer := image.NewNRGBA(image.Rect(0, 0, W, H))

	for py := 0; py < H; py++ {
		// Latitude is constant across a row, but the subsolar hour angle varies
		// with longitude, so elevation must be computed per pixel. We still hoist
		// the row latitude.
		_, lat := c.Grid.PxToLonLat(0, py)
		rowOff := layer.PixOffset(0, py)
		for px := 0; px < W; px++ {
			lon := (float64(px)+0.5)/float64(W)*360.0 - 180.0
			elev := pos.Elevation(lon, lat)

			var r, g, b, a uint8
			if elev > 0 {
				ghi := solar.GlobalHorizontalIrradiance(elev, month)
				norm := ghi / maxGHI
				if norm > 1 {
					norm = 1
				}
				r, g, b = cmap(norm)
				a = uint8(clamp01(sc.DayAlpha*norm) * 255)
			} else {
				f := clamp01(-elev / twilight)
				a = uint8(clamp01(darkness*f) * 255)
				// r,g,b stay 0 => black night shading
			}
			o := rowOff + px*4
			layer.Pix[o+0] = r
			layer.Pix[o+1] = g
			layer.Pix[o+2] = b
			layer.Pix[o+3] = a
		}
	}
	draw.Draw(c.Img, c.Img.Bounds(), layer, image.Point{}, draw.Over)
}
