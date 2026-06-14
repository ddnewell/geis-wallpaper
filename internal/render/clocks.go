package render

import (
	"encoding/json"
	"image"
	"image/color"
	"os"
	"sort"
	"time"

	"github.com/ddnewell/geis-wallpaper/internal/solar"
)

// clockDef is one city in clocks.json.
type clockDef struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
	TZ  string  `json:"tz"`
}

type clocksFile struct {
	Formatting struct {
		Display struct {
			Lat    float64 `json:"lat"`
			Offset float64 `json:"offset"`
		} `json:"display"`
		Colors []string `json:"colors"`
	} `json:"formatting"`
	Clocks map[string]clockDef `json:"clocks"`
}

// bandTopLat matches the legacy clock band top edge.
const bandTopLat = -64.25

// DrawClocks renders the world-clock band and city labels. Pure local
// computation (tz + solar altitude); never touches the network. A missing or
// invalid file is a no-op so the render still succeeds.
func DrawClocks(c *Canvas, clocksPath string, t time.Time) {
	data, err := os.ReadFile(clocksPath)
	if err != nil {
		return
	}
	var cf clocksFile
	if err := json.Unmarshal(data, &cf); err != nil || len(cf.Clocks) == 0 {
		return
	}

	// Translucent wheat band across the bottom.
	_, yTop := c.Grid.LonLatToXY(0, bandTopLat)
	c.FillRect(image.Rect(0, int(yTop), c.Grid.W, c.Grid.H), nrgba(245, 222, 179, 120))

	dlat := cf.Formatting.Display.Lat
	if dlat == 0 {
		dlat = -74
	}
	offset := cf.Formatting.Display.Offset
	if offset == 0 {
		offset = 15
	}
	palette := parseHexColors(cf.Formatting.Colors)
	if len(palette) == 0 {
		palette = []color.NRGBA{{R: 200, G: 200, B: 200, A: 255}}
	}

	// Sort cities by longitude (west to east).
	type city struct {
		code string
		def  clockDef
	}
	cities := make([]city, 0, len(cf.Clocks))
	for code, def := range cf.Clocks {
		cities = append(cities, city{code, def})
	}
	sort.Slice(cities, func(i, j int) bool { return cities[i].def.Lon < cities[j].def.Lon })

	pos := solar.SubsolarPoint(t)
	prevLon := -1000.0
	up := false

	for i, ct := range cities {
		// De-overlap: alternate above/below when too close to the previous city.
		if i > 0 && ct.def.Lon-prevLon < offset {
			up = !up
		} else {
			up = false
		}
		prevLon = ct.def.Lon

		x, baseY := c.Grid.LonLatToXY(ct.def.Lon, dlat)
		_, cityY := c.Grid.LonLatToXY(ct.def.Lon, ct.def.Lat)
		px, pBaseY, pCityY := int(x), int(baseY), int(cityY)

		dotCol := palette[i%len(palette)]
		// City location marker and band reference dot.
		DrawDisc(c.Img, px, pCityY, 5, dotCol)
		DrawDisc(c.Img, px, pBaseY, 4, dotCol)

		// Local time string.
		clockStr := "--:--"
		if loc, err := time.LoadLocation(ct.def.TZ); err == nil {
			clockStr = t.In(loc).Format("15:04")
		}

		// Label color by sun altitude (legacy bands).
		alt := pos.Elevation(ct.def.Lon, ct.def.Lat)
		txtCol := altitudeColor(alt)

		var codeY, timeY int
		if up {
			codeY, timeY = pBaseY-26, pBaseY-6
		} else {
			codeY, timeY = pBaseY+18, pBaseY+40
		}
		DrawLabel(c.Img, px, codeY, 22, true, txtCol, AnchorCenter, ct.code)
		DrawLabel(c.Img, px, timeY, 22, false, txtCol, AnchorCenter, clockStr)
	}
}

// altitudeColor reproduces the legacy 4-band day/twilight/night label colors.
func altitudeColor(alt float64) color.NRGBA {
	switch {
	case alt < -8:
		return color.NRGBA{R: 0x09, G: 0x04, B: 0x1c, A: 255}
	case alt < -2:
		return color.NRGBA{R: 0x22, G: 0x1c, B: 0x32, A: 255}
	case alt < 5:
		return color.NRGBA{R: 0x1a, G: 0x06, B: 0x62, A: 255}
	default:
		return color.NRGBA{R: 0x12, G: 0x57, B: 0x00, A: 255}
	}
}

func parseHexColors(hexes []string) []color.NRGBA {
	out := make([]color.NRGBA, 0, len(hexes))
	for _, h := range hexes {
		if c, ok := parseHex(h); ok {
			out = append(out, c)
		}
	}
	return out
}

func parseHex(h string) (color.NRGBA, bool) {
	if len(h) == 7 && h[0] == '#' {
		r, ok1 := hexByte(h[1:3])
		g, ok2 := hexByte(h[3:5])
		b, ok3 := hexByte(h[5:7])
		if ok1 && ok2 && ok3 {
			return color.NRGBA{R: r, G: g, B: b, A: 255}, true
		}
	}
	return color.NRGBA{}, false
}

func hexByte(s string) (uint8, bool) {
	var v int
	for _, ch := range s {
		v <<= 4
		switch {
		case ch >= '0' && ch <= '9':
			v |= int(ch - '0')
		case ch >= 'a' && ch <= 'f':
			v |= int(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			v |= int(ch-'A') + 10
		default:
			return 0, false
		}
	}
	return uint8(v), true
}
