package render

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"path/filepath"
	"time"

	"github.com/ddnewell/geis-wallpaper/internal/config"
	"github.com/ddnewell/geis-wallpaper/internal/feeds"
	"github.com/ddnewell/geis-wallpaper/internal/geo"
	xdraw "golang.org/x/image/draw"
)

// Master canvas dimensions: the 2:1 equirectangular grid matching the base map.
const (
	MasterW = 2560
	MasterH = 1280
)

// maxOverlayAge drops event overlays older than this (e.g. a long offline
// stretch) so the map never shows badly stale positions. Solar + clocks always
// render regardless.
const maxOverlayAge = 24 * time.Hour

// Inputs carries everything the renderer needs. All overlay fields are optional;
// a zero/nil value simply omits that layer, which is how the fully-offline
// fresh-install case still produces a complete wallpaper.
type Inputs struct {
	Cfg        config.Config
	Now        time.Time
	AssetDir   string
	ClocksFile string

	// Cached event data (nil if never fetched).
	Payload    *feeds.Payload
	PayloadAge time.Duration

	// Cached cloud/satellite image (nil if none).
	Cloud    image.Image
	CloudAge time.Duration
}

// RenderMaster composites all enabled layers at master resolution. It performs
// no network I/O.
func RenderMaster(in Inputs) (*Canvas, error) {
	c := NewCanvas(MasterW, MasterH)

	// 1. Base map.
	basePath := filepath.Join(in.AssetDir, "img", "NaturalEarth_Mac13Retina.png")
	base, err := LoadBaseMap(basePath, c.Grid)
	if err != nil {
		return nil, err
	}
	c.DrawOver(base)

	overlaysFresh := in.Payload != nil && in.PayloadAge <= maxOverlayAge

	// 2. Clouds below the solar layer (so night clouds darken too), unless
	// configured otherwise.
	drawClouds := in.Cfg.Overlays.Satellite && in.Cloud != nil && in.CloudAge <= maxOverlayAge
	if drawClouds && !in.Cfg.SatelliteAboveSolar {
		c.DrawAlpha(in.Cloud, in.Cfg.Satellite.Alpha)
	}

	// 3. Solar irradiance heatmap + night shading.
	DrawSolar(c, in.Now, in.Cfg.Solar, in.Cfg.Darkness)

	if drawClouds && in.Cfg.SatelliteAboveSolar {
		c.DrawAlpha(in.Cloud, in.Cfg.Satellite.Alpha)
	}

	// 4-5. Event overlays.
	if overlaysFresh {
		p := in.Payload
		if in.Cfg.Overlays.Aurora {
			DrawAurora(c, p.Aurora)
		}
		if in.Cfg.Overlays.Storms {
			DrawStorms(c, p.Storms, in.AssetDir)
		}
		if in.Cfg.Overlays.Earthquakes {
			DrawQuakes(c, p.Quakes)
		}
		if in.Cfg.Overlays.Fires {
			DrawFires(c, p.Fires)
		}
		if in.Cfg.Overlays.Ships {
			DrawShips(c, p.Ships)
		}
		if in.Cfg.Overlays.ISS {
			DrawISS(c, p.ISS)
		}
	}

	// 6. World clocks.
	DrawClocks(c, in.ClocksFile, in.Now)

	// 7. Footer: render time + overlay staleness.
	DrawFooter(c, in)

	return c, nil
}

// DrawFooter writes the "Updated" line plus an overlay-staleness note.
func DrawFooter(c *Canvas, in Inputs) {
	red := color.NRGBA{R: 0xa6, A: 235}
	DrawLabel(c.Img, 28, c.Grid.H-22, 26, true, red, AnchorLeft,
		"Updated "+in.Now.Format("Jan 2, 2006  3:04 PM"))

	var note string
	switch {
	case in.Payload == nil:
		note = "Events: no data yet"
	case in.PayloadAge > maxOverlayAge:
		note = "Events: cached " + humanizeAge(in.PayloadAge) + " ago (too stale — hidden)"
	case in.PayloadAge > 30*time.Minute:
		note = "Events: cached " + humanizeAge(in.PayloadAge) + " ago"
	default:
		note = "Events: live (" + humanizeAge(in.PayloadAge) + " ago)"
	}
	DrawLabel(c.Img, c.Grid.W-28, c.Grid.H-22, 22, false, red, AnchorRight, note)
}

func humanizeAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "moments"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// CropToScreen scales the master up (Catmull-Rom) if the screen exceeds it in
// either dimension, then center-crops to (sw,sh).
func CropToScreen(master *image.RGBA, sw, sh int) *image.RGBA {
	if sw <= 0 || sh <= 0 {
		sw, sh = MasterW, MasterH
	}
	mw, mh := master.Bounds().Dx(), master.Bounds().Dy()
	scale := math.Max(float64(sw)/float64(mw), float64(sh)/float64(mh))
	if scale > 1 {
		nw := int(math.Ceil(float64(mw) * scale))
		nh := int(math.Ceil(float64(mh) * scale))
		scaled := image.NewRGBA(image.Rect(0, 0, nw, nh))
		xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), master, master.Bounds(), xdraw.Over, nil)
		master, mw, mh = scaled, nw, nh
	}
	rect := geo.CenterCropRect(mw, mh, sw, sh)
	out := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	xdraw.Copy(out, image.Point{}, master, rect, xdraw.Src, nil)
	return out
}
