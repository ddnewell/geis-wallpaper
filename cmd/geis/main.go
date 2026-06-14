// Command geis is the GEIS wallpaper renderer. It runs entirely offline: it
// reads the base map, clocks, and any locally-cached overlay data, composites
// the present-time solar irradiance heatmap, writes a uniquely-named PNG, and
// sets it as the desktop wallpaper. It is intended to be run on an interval by
// a launchd LaunchAgent.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "time/tzdata" // embed the tz database so world clocks work offline

	"github.com/ddnewell/geis-wallpaper/internal/cache"
	"github.com/ddnewell/geis-wallpaper/internal/config"
	"github.com/ddnewell/geis-wallpaper/internal/feeds"
	"github.com/ddnewell/geis-wallpaper/internal/platform"
	"github.com/ddnewell/geis-wallpaper/internal/render"
)

func main() {
	cfgPath := flag.String("config", "", "path to config.json")
	dryRun := flag.Bool("dry-run", false, "render and write the frame but do not set the wallpaper")
	flag.Parse()

	log.SetFlags(log.LstdFlags)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	outDir := cfg.OutputDir()
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatalf("create frames dir: %v", err)
	}

	// Serialize against overlapping launchd ticks.
	lock, err := platform.TryLock(filepath.Join(outDir, ".lock"))
	if err != nil {
		log.Fatalf("lock: %v", err)
	}
	if lock == nil {
		log.Printf("another geis run is in progress; skipping this cycle")
		return
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Printf("unlock: %v", err)
		}
	}()

	now := time.Now()

	in := render.Inputs{
		Cfg:        cfg,
		Now:        now,
		AssetDir:   cfg.AssetDir(),
		ClocksFile: cfg.ClocksFile(),
	}
	loadCached(&in, cfg, now)

	c, err := render.RenderMaster(in)
	if err != nil {
		log.Fatalf("render: %v", err)
	}

	// Per-display cropping: one image tailored to each screen's aspect, or a
	// single image for all displays (the default). Each target writes to a
	// STABLE path (overwritten in place); RefreshSpaces then forces every Space
	// to re-read it, which both bypasses the same-path cache and updates all
	// desktops rather than only the current Space.
	targets := renderTargets(cfg)
	keep := make(map[string]bool, len(targets))
	wallpaperSet := false
	for _, t := range targets {
		out := render.CropToScreen(c.Img, t.w, t.h)
		name := "wallpaper.png"
		if t.screen >= 0 {
			name = fmt.Sprintf("wallpaper-%d.png", t.screen)
		}
		framePath, err := platform.WriteStable(outDir, name, out)
		if err != nil {
			log.Fatalf("write frame: %v", err)
		}
		keep[framePath] = true
		if *dryRun {
			log.Printf("dry-run: wrote %s (%dx%d)%s", framePath, t.w, t.h, t.label)
			continue
		}
		if t.screen < 0 {
			err = platform.SetWallpaper(framePath) // every display, no TCC prompt under launchd
		} else {
			err = platform.SetWallpaperForScreen(t.screen, framePath)
		}
		if err != nil {
			log.Printf("set wallpaper%s: %v", t.label, err)
		} else {
			wallpaperSet = true
			log.Printf("wallpaper updated: %s (%dx%d)%s", framePath, t.w, t.h, t.label)
		}
	}

	if !*dryRun {
		if wallpaperSet {
			// Force every Space to reload the (in-place updated) image.
			_ = platform.RefreshSpaces()
		}
		platform.CleanupFrames(outDir, keep)
	}
}

type target struct {
	screen int // -1 = all screens
	w, h   int
	label  string
}

// renderTargets decides what to render and where to set it.
func renderTargets(cfg config.Config) []target {
	if cfg.Displays.RenderPerDisplay {
		if n := platform.ScreenCount(); n > 1 {
			ts := make([]target, 0, n)
			for i := 0; i < n; i++ {
				w, h, err := platform.ScreenSize(i)
				if err != nil || w <= 0 || h <= 0 {
					w, h = render.MasterW, render.MasterH
				}
				ts = append(ts, target{screen: i, w: w, h: h, label: fmt.Sprintf(" [display %d]", i)})
			}
			return ts
		}
	}
	w, h := screenSize(cfg)
	return []target{{screen: -1, w: w, h: h}}
}

// loadCached reads the locally-cached payload and cloud image into the render
// inputs. Missing/corrupt caches are ignored (overlays simply omitted) so the
// render always succeeds offline.
func loadCached(in *render.Inputs, cfg config.Config, now time.Time) {
	store := cache.New(cfg.CacheDir())
	if data, _, ok := store.Read("payload.json"); ok {
		var p feeds.Payload
		if err := json.Unmarshal(data, &p); err == nil {
			in.Payload = &p
			if age, ok := store.Age("payload.json", now); ok {
				in.PayloadAge = age
			}
		}
	}
	if data, _, ok := store.Read("satellite.png"); ok {
		if img, _, err := image.Decode(bytes.NewReader(data)); err == nil {
			in.Cloud = img
			if age, ok := store.Age("satellite.png", now); ok {
				in.CloudAge = age
			}
		}
	}
}

// screenSize resolves the output dimensions: auto-detected main display,
// configured screen, or the 2:1 master fallback.
func screenSize(cfg config.Config) (int, int) {
	if cfg.Displays.AutoDetect {
		if w, h, err := platform.MainScreenSize(); err == nil && w > 0 && h > 0 {
			return w, h
		}
	}
	if len(cfg.Displays.Screens) > 0 {
		s := cfg.Displays.Screens[0]
		if s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}
	return render.MasterW, render.MasterH
}
