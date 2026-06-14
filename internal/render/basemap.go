package render

import (
	"fmt"
	"image"
	_ "image/jpeg" // decoder registration
	_ "image/png"  // decoder registration
	"os"

	"github.com/ddnewell/geis-wallpaper/internal/geo"
	xdraw "golang.org/x/image/draw"
)

// LoadBaseMap decodes the base map PNG and returns it as an *image.RGBA sized
// exactly to the grid (scaled with Catmull-Rom if the source differs).
func LoadBaseMap(path string, g geo.Grid) (*image.RGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open base map: %w", err)
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode base map: %w", err)
	}
	return toGridRGBA(src, g), nil
}

// toGridRGBA returns src as an RGBA image of exactly g.W x g.H pixels.
func toGridRGBA(src image.Image, g geo.Grid) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, g.W, g.H))
	if src.Bounds().Dx() == g.W && src.Bounds().Dy() == g.H {
		xdraw.Copy(dst, image.Point{}, src, src.Bounds(), xdraw.Src, nil)
	} else {
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	}
	return dst
}
