package geo

import (
	"math"
	"testing"
)

func TestLonLatToXY(t *testing.T) {
	g := Grid{W: 2560, H: 1280}
	cases := []struct {
		lon, lat float64
		wantX    float64
		wantY    float64
		desc     string
	}{
		{-180, 90, 0, 0, "NW corner -> (0,0)"},
		{180, -90, 2560, 1280, "SE corner -> (W,H)"},
		{0, 0, 1280, 640, "origin -> center"},
		{0, 90, 1280, 0, "north pole -> top center (y=0, NOT H)"},
		{0, -90, 1280, 1280, "south pole -> bottom center"},
	}
	for _, c := range cases {
		x, y := g.LonLatToXY(c.lon, c.lat)
		if math.Abs(x-c.wantX) > 1e-9 || math.Abs(y-c.wantY) > 1e-9 {
			t.Errorf("%s: LonLatToXY(%v,%v)=(%v,%v) want (%v,%v)", c.desc, c.lon, c.lat, x, y, c.wantX, c.wantY)
		}
	}
}

// TestRoundTrip ensures PxToLonLat is the inverse of LonLatToPx (within a pixel).
func TestRoundTrip(t *testing.T) {
	g := Grid{W: 2560, H: 1280}
	for _, px := range []int{0, 100, 1280, 2559} {
		for _, py := range []int{0, 100, 640, 1279} {
			lon, lat := g.PxToLonLat(px, py)
			gpx, gpy := g.LonLatToPx(lon, lat)
			if gpx != px || gpy != py {
				t.Errorf("round trip px(%d,%d)->lonlat(%v,%v)->px(%d,%d)", px, py, lon, lat, gpx, gpy)
			}
		}
	}
}

// TestYFlip guards the most common porting bug: north must map to a SMALLER y
// than south.
func TestYFlip(t *testing.T) {
	g := Grid{W: 2560, H: 1280}
	_, yNorth := g.LonLatToXY(0, 45)
	_, ySouth := g.LonLatToXY(0, -45)
	if !(yNorth < ySouth) {
		t.Fatalf("expected north (y=%v) above south (y=%v) with top-left origin", yNorth, ySouth)
	}
}
