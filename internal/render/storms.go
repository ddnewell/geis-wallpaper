package render

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/ddnewell/geis-wallpaper/internal/feeds"
)

// tropicalIcon maps Saffir-Simpson category (-5..5) to the icon filename,
// matching the legacy wmap.py tropicalImgPath and the assets in
// ico/wx/tropical/{normal,small}/.
var tropicalIcon = map[int]string{
	-5: "remnants.png",
	-4: "invest.png",
	-3: "extratropical.png",
	-2: "depression.png",
	-1: "depression.png",
	0:  "tropical-storm.png",
	1:  "hurricane-1.png",
	2:  "hurricane-2.png",
	3:  "hurricane-3.png",
	4:  "hurricane-4.png",
	5:  "hurricane-5.png",
}

var (
	iconMu    sync.Mutex
	iconCache = map[string]image.Image{}
)

func loadIcon(assetDir, size string, cat int) image.Image {
	name, ok := tropicalIcon[cat]
	if !ok {
		name = tropicalIcon[0]
	}
	path := filepath.Join(assetDir, "ico", "wx", "tropical", size, name)
	iconMu.Lock()
	defer iconMu.Unlock()
	if img, seen := iconCache[path]; seen {
		return img
	}
	var img image.Image
	if f, err := os.Open(path); err == nil {
		if decoded, _, derr := image.Decode(f); derr == nil {
			img = decoded
		}
		f.Close()
	}
	iconCache[path] = img // cache nil too, to avoid re-stat on every storm
	return img
}

// DrawStorms plots active tropical cyclones: faint forecast track icons first,
// then the current-position icon and storm name. Reuses ico/wx/tropical/.
func DrawStorms(c *Canvas, storms []feeds.Storm, assetDir string) {
	for _, s := range storms {
		// Forecast track (small icons, faint).
		for _, f := range s.Forecast {
			if ic := loadIcon(assetDir, "small", f.Category); ic != nil {
				pasteCentered(c.Img, ic, c, f.Lon, f.Lat, 0.4)
			}
		}
		// Current position (normal icon).
		x, y := c.Grid.LonLatToXY(s.Lon, s.Lat)
		if ic := loadIcon(assetDir, "normal", s.Category); ic != nil {
			pasteCentered(c.Img, ic, c, s.Lon, s.Lat, 0.85)
			DrawLabel(c.Img, int(x), int(y)+ic.Bounds().Dy()/2+18, 20, true,
				color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 230}, AnchorCenter, s.Name)
		}
	}
}

// pasteCentered composites src centered on (lon,lat) with a global alpha,
// respecting src's own transparency.
func pasteCentered(dst *image.RGBA, src image.Image, c *Canvas, lon, lat, alpha float64) {
	x, y := c.Grid.LonLatToXY(lon, lat)
	b := src.Bounds()
	x0 := int(x) - b.Dx()/2
	y0 := int(y) - b.Dy()/2
	mask := image.NewUniform(color.Alpha{A: uint8(clamp01(alpha) * 255)})
	dr := image.Rect(x0, y0, x0+b.Dx(), y0+b.Dy())
	draw.DrawMask(dst, dr, src, b.Min, mask, image.Point{}, draw.Over)
}
