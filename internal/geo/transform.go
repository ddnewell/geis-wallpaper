// Package geo provides the equirectangular (PlateCarree) lon/lat <-> pixel
// transform for a full-globe map (-180..180 lon, -90..90 lat).
//
// The original Python used cartopy for this, but the projection is a single
// linear map. The one subtlety is the origin: matplotlib used a bottom-left
// origin (cy = (lat+90)/180*H), whereas Go's image package uses a TOP-LEFT
// origin, so latitude increases UPWARD => y = (90-lat)/180*H. Getting this
// wrong mirrors the map vertically.
package geo

import (
	"image"
	"math"
)

// Grid is the master canvas dimensions in pixels for the full globe.
type Grid struct {
	W, H int
}

// LonLatToXY maps geographic coordinates to floating-point pixel coordinates
// on the grid, with a top-left origin (y grows southward).
func (g Grid) LonLatToXY(lon, lat float64) (x, y float64) {
	x = (lon + 180.0) / 360.0 * float64(g.W)
	y = (90.0 - lat) / 180.0 * float64(g.H)
	return x, y
}

// LonLatToPx maps geographic coordinates to integer pixel indices, clamped to
// the grid bounds.
func (g Grid) LonLatToPx(lon, lat float64) (px, py int) {
	x, y := g.LonLatToXY(lon, lat)
	px = clampInt(int(math.Floor(x)), 0, g.W-1)
	py = clampInt(int(math.Floor(y)), 0, g.H-1)
	return px, py
}

// PxToLonLat maps the CENTER of integer pixel (px,py) back to lon/lat. Used to
// evaluate per-pixel fields (e.g. solar elevation) across the whole grid.
func (g Grid) PxToLonLat(px, py int) (lon, lat float64) {
	lon = (float64(px)+0.5)/float64(g.W)*360.0 - 180.0
	lat = 90.0 - (float64(py)+0.5)/float64(g.H)*180.0
	return lon, lat
}

// CenterCropRect returns the centered rectangle of size (sw,sh) within a canvas
// of size (cw,ch). It assumes the canvas is at least as large as the crop in
// both dimensions; callers should scale the canvas up first if it is not.
// Ports the crop math from the legacy wmap.py set_wallpaper().
func CenterCropRect(cw, ch, sw, sh int) image.Rectangle {
	if sw > cw {
		sw = cw
	}
	if sh > ch {
		sh = ch
	}
	x0 := (cw - sw) / 2
	y0 := (ch - sh) / 2
	return image.Rect(x0, y0, x0+sw, y0+sh)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
