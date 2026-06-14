package render

import (
	"image"
	"image/color"

	"github.com/ddnewell/geis-wallpaper/internal/fonts"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Anchor controls horizontal text alignment relative to the given x.
type Anchor int

const (
	AnchorLeft Anchor = iota
	AnchorCenter
	AnchorRight
)

// DrawString draws text onto dst at baseline (x,y) with the given face, color,
// and horizontal anchor.
func DrawString(dst *image.RGBA, face font.Face, col color.Color, x, y int, anchor Anchor, text string) {
	if face == nil {
		return
	}
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
	}
	w := d.MeasureString(text).Round()
	switch anchor {
	case AnchorCenter:
		x -= w / 2
	case AnchorRight:
		x -= w
	}
	d.Dot = fixed.P(x, y)
	d.DrawString(text)
}

// DrawLabel draws text at the given size/weight; convenience over DrawString.
func DrawLabel(dst *image.RGBA, x, y int, size float64, bold bool, col color.Color, anchor Anchor, text string) {
	DrawString(dst, fonts.Face(size, bold), col, x, y, anchor, text)
}
