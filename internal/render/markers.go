package render

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/ddnewell/geis-wallpaper/internal/feeds"
)

// DrawQuakes plots earthquakes as discs sized by magnitude and colored by depth
// (shallow = red/orange, deep = blue), replacing the matplotlib scatter.
func DrawQuakes(c *Canvas, quakes []feeds.Quake) {
	for _, q := range quakes {
		x, y := c.Grid.LonLatToXY(q.Lon, q.Lat)
		r := int(3 + (q.Mag-3)*2.2)
		if r < 3 {
			r = 3
		}
		if r > 40 {
			r = 40
		}
		col := depthColor(q.DepthKm)
		// Soft halo then solid core.
		DrawDisc(c.Img, int(x), int(y), r+2, color.NRGBA{R: col.R, G: col.G, B: col.B, A: 70})
		DrawDisc(c.Img, int(x), int(y), r, color.NRGBA{R: col.R, G: col.G, B: col.B, A: 200})
	}
}

// depthColor maps earthquake depth (km) to a hazard-style color ramp.
func depthColor(depthKm float64) color.NRGBA {
	switch {
	case depthKm < 70:
		return color.NRGBA{R: 0xe6, G: 0x39, B: 0x2b, A: 255} // shallow: red
	case depthKm < 300:
		return color.NRGBA{R: 0xf5, G: 0xa6, B: 0x23, A: 255} // intermediate: amber
	default:
		return color.NRGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 255} // deep: blue
	}
}

// DrawAurora composites the OVATION aurora oval as soft green cells whose
// opacity scales with forecast intensity (probability). Built in one layer so
// the ~thousands of cells cost a single composite, not thousands of draws.
func DrawAurora(c *Canvas, cells []feeds.AuroraCell) {
	if len(cells) == 0 {
		return
	}
	W, H := c.Grid.W, c.Grid.H
	layer := image.NewNRGBA(image.Rect(0, 0, W, H))
	// Each 1-degree cell spans ~W/360 px; draw a small block so the oval reads
	// as a continuous band.
	half := W / 720 // ~half a degree
	if half < 2 {
		half = 2
	}
	for _, cell := range cells {
		px, py := c.Grid.LonLatToPx(cell.Lon, cell.Lat)
		a := cell.Intensity * 5 // 0..100 probability -> alpha scale
		if a > 220 {
			a = 220
		}
		col := color.NRGBA{R: 60, G: 240, B: 130, A: uint8(a)}
		for dy := -half; dy <= half; dy++ {
			yy := py + dy
			if yy < 0 || yy >= H {
				continue
			}
			for dx := -half; dx <= half; dx++ {
				xx := px + dx
				if xx < 0 || xx >= W {
					continue
				}
				// Keep the brightest contribution per pixel.
				if layer.NRGBAAt(xx, yy).A < col.A {
					layer.SetNRGBA(xx, yy, col)
				}
			}
		}
	}
	draw.Draw(c.Img, c.Img.Bounds(), layer, image.Point{}, draw.Over)
}

// DrawShips plots AIS vessel positions as small diamonds (legacy color).
func DrawShips(c *Canvas, ships []feeds.Ship) {
	col := color.NRGBA{R: 0xad, G: 0xce, B: 0xfa, A: 255}
	for _, s := range ships {
		x, y := c.Grid.LonLatToXY(s.Lon, s.Lat)
		DrawDiamond(c.Img, int(x), int(y), 4, col)
	}
}

// DrawFires plots NASA FIRMS active-fire detections as small hot dots.
func DrawFires(c *Canvas, fires []feeds.Fire) {
	col := color.NRGBA{R: 0xff, G: 0x55, B: 0x10, A: 210}
	for _, f := range fires {
		x, y := c.Grid.LonLatToXY(f.Lon, f.Lat)
		DrawDisc(c.Img, int(x), int(y), 3, col)
	}
}

// DrawISS plots the ISS position with a cyan marker and label.
func DrawISS(c *Canvas, iss *feeds.ISS) {
	if iss == nil {
		return
	}
	x, y := c.Grid.LonLatToXY(iss.Lon, iss.Lat)
	cyan := color.NRGBA{R: 0x6e, G: 0xe7, B: 0xf5, A: 255}
	DrawDiamond(c.Img, int(x), int(y), 9, color.NRGBA{R: 0x6e, G: 0xe7, B: 0xf5, A: 90})
	DrawDiamond(c.Img, int(x), int(y), 6, cyan)
	DrawLabel(c.Img, int(x), int(y)-14, 20, true, cyan, AnchorCenter, "ISS")
}
