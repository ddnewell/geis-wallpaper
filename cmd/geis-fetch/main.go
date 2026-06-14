// Command geis-fetch pulls event data into the local cache. It is deliberately
// SEPARATE from the renderer: a slow or failing fetch can never block or break
// the wallpaper update. Each feed falls back to its previously-cached value, so
// a transient failure dims nothing. Run on a longer interval than the renderer.
//
// When a companion server is configured it is the preferred source (one request
// for everything, including the keyed ships/fires). If the companion is
// unreachable, the client falls back to fetching the keyless feeds directly, so
// the wallpaper keeps getting fresh storms/quakes/aurora/iss/clouds regardless.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/ddnewell/geis-wallpaper/internal/cache"
	"github.com/ddnewell/geis-wallpaper/internal/config"
	"github.com/ddnewell/geis-wallpaper/internal/feeds"
)

func main() {
	cfgPath := flag.String("config", "", "path to config.json")
	flag.Parse()
	log.SetFlags(log.LstdFlags)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	store := cache.New(cfg.CacheDir())
	now := time.Now()

	var prev feeds.Payload
	if data, _, ok := store.Read("payload.json"); ok {
		_ = json.Unmarshal(data, &prev)
	}
	p := feeds.Payload{GeneratedAt: now}

	companionUsed := false
	if cfg.Companion.Enabled && cfg.Companion.BaseURL != "" {
		companionUsed = fetchFromCompanion(cfg, store, now, &p)
	}
	if !companionUsed {
		fetchDirect(cfg, store, now, prev, &p)
	}

	data, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("marshal payload: %v", err)
	}
	if err := store.Write("payload.json", data, now); err != nil {
		log.Fatalf("write payload: %v", err)
	}
	log.Printf("payload cached: %d bytes (storms=%d quakes=%d aurora=%d iss=%v ships=%d fires=%d) companion=%v",
		len(data), len(p.Storms), len(p.Quakes), len(p.Aurora), p.ISS != nil, len(p.Ships), len(p.Fires), companionUsed)
}

// fetchFromCompanion pulls the aggregated payload + satellite from the companion.
// Returns false (so the caller falls back to direct) if the payload fetch fails.
func fetchFromCompanion(cfg config.Config, store *cache.Store, now time.Time, p *feeds.Payload) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cp, err := feeds.CompanionPayload(ctx, cfg.Companion.BaseURL)
	if err != nil {
		log.Printf("companion payload: FAILED (%v) — falling back to direct feeds", err)
		return false
	}
	*p = *cp
	p.GeneratedAt = now
	log.Printf("companion payload: ok")

	if cfg.Overlays.Satellite {
		run("companion satellite", func(ctx context.Context) error {
			img, err := feeds.CompanionSatellite(ctx, cfg.Companion.BaseURL)
			if err != nil {
				return err
			}
			return store.Write("satellite.png", img, now)
		})
	}
	return true
}

// fetchDirect fetches the keyless feeds straight from their sources. Ships and
// fires require keys and are only available via the companion, so they carry
// over from the previous payload here.
func fetchDirect(cfg config.Config, store *cache.Store, now time.Time, prev feeds.Payload, p *feeds.Payload) {
	if cfg.Overlays.Storms {
		run("storms", func(ctx context.Context) error {
			s, err := feeds.FetchStorms(ctx)
			if err != nil {
				p.Storms = prev.Storms
				return err
			}
			p.Storms = s
			return nil
		})
	}
	if cfg.Overlays.Earthquakes {
		run("quakes", func(ctx context.Context) error {
			q, err := feeds.FetchQuakes(ctx, cfg.Earthquakes.MinMagnitude)
			if err != nil {
				p.Quakes = prev.Quakes
				return err
			}
			p.Quakes = q
			return nil
		})
	}
	if cfg.Overlays.Aurora {
		run("aurora", func(ctx context.Context) error {
			a, err := feeds.FetchAurora(ctx)
			if err != nil {
				p.Aurora = prev.Aurora
				return err
			}
			p.Aurora = a
			return nil
		})
	}
	if cfg.Overlays.ISS {
		run("iss", func(ctx context.Context) error {
			iss, err := feeds.FetchISS(ctx)
			if err != nil {
				p.ISS = prev.ISS
				return err
			}
			p.ISS = iss
			return nil
		})
	}
	if cfg.Overlays.Satellite {
		run("satellite", func(ctx context.Context) error {
			img, err := feeds.FetchSatellite(ctx, cfg.Satellite.Layer, now)
			if err != nil {
				return err
			}
			return store.Write("satellite.png", img, now)
		})
	}
	// Keyed feeds: no client keys, so preserve whatever the companion last gave.
	p.Ships = prev.Ships
	p.Fires = prev.Fires
}

// run executes a single fetch with its own timeout and logs the outcome.
func run(name string, fn func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := fn(ctx); err != nil {
		log.Printf("%s: FAILED (%v) — kept last-good", name, err)
		return
	}
	log.Printf("%s: ok", name)
}
