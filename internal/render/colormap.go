package render

import "image/color"

// Colormap maps a normalized value t in [0,1] to an RGB color (alpha handled by
// the caller).
type Colormap func(t float64) (r, g, b uint8)

type stop struct {
	t       float64
	r, g, b uint8
}

func gradient(stops []stop) Colormap {
	return func(t float64) (uint8, uint8, uint8) {
		t = clamp01(t)
		for i := 1; i < len(stops); i++ {
			if t <= stops[i].t {
				a, b := stops[i-1], stops[i]
				span := b.t - a.t
				f := 0.0
				if span > 0 {
					f = (t - a.t) / span
				}
				return lerp8(a.r, b.r, f), lerp8(a.g, b.g, f), lerp8(a.b, b.b, f)
			}
		}
		last := stops[len(stops)-1]
		return last.r, last.g, last.b
	}
}

// solarMap: a warm "radiation" gradient (dim red -> orange -> amber -> hot white).
var solarStops = []stop{
	{0.00, 60, 10, 50},
	{0.20, 150, 35, 45},
	{0.45, 225, 95, 30},
	{0.70, 250, 170, 45},
	{1.00, 255, 244, 190},
}

// inferno: perceptually-ordered hot map approximation.
var infernoStops = []stop{
	{0.00, 0, 0, 4},
	{0.25, 87, 16, 110},
	{0.50, 188, 55, 84},
	{0.75, 249, 142, 9},
	{1.00, 252, 255, 164},
}

// Colormaps returns the named colormap, defaulting to "solar".
func Colormaps(name string) Colormap {
	switch name {
	case "inferno":
		return gradient(infernoStops)
	case "grayscale":
		return func(t float64) (uint8, uint8, uint8) {
			v := uint8(clamp01(t) * 255)
			return v, v, v
		}
	default:
		return gradient(solarStops)
	}
}

func lerp8(a, b uint8, f float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*f)
}

// nrgba is a small helper to build an NRGBA color.
func nrgba(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }
