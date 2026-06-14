// Command geisd is the GEIS companion server, designed for DreamHost shared
// hosting. It has two jobs:
//
//   - `geisd -refresh`: run by cron (~10 min). Fetches and normalizes ALL feeds
//     — including the keyed AISHub (ships) and FIRMS (fires), whose secrets live
//     in the config file OUTSIDE the web root — and writes static
//     payload.json / satellite.png / ships.json into the data dir.
//   - `geisd -fcgi` (the dispatch.fcgi handler) or `geisd -http :PORT` (dev):
//     serves those static files. No upstream calls happen in the request path,
//     so responses are always fast.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ddnewell/geis-wallpaper/internal/cache"
	"github.com/ddnewell/geis-wallpaper/internal/feeds"
)

type serverConfig struct {
	DataDir   string `json:"data_dir"`
	Satellite struct {
		Layer string `json:"layer"`
	} `json:"satellite"`
	Earthquakes struct {
		MinMagnitude float64 `json:"min_magnitude"`
	} `json:"earthquakes"`
	Overlays struct {
		Satellite   bool `json:"satellite"`
		Storms      bool `json:"storms"`
		Earthquakes bool `json:"earthquakes"`
		Aurora      bool `json:"aurora"`
		ISS         bool `json:"iss"`
		Ships       bool `json:"ships"`
		Fires       bool `json:"fires"`
	} `json:"overlays"`
	AISHub struct {
		Username string  `json:"username"`
		LatMin   float64 `json:"lat_min"`
		LatMax   float64 `json:"lat_max"`
		LonMin   float64 `json:"lon_min"`
		LonMax   float64 `json:"lon_max"`
	} `json:"aishub"`
	FIRMS struct {
		MapKey   string  `json:"map_key"`
		Source   string  `json:"source"`
		Area     string  `json:"area"`
		DayRange int     `json:"day_range"`
		MinFRP   float64 `json:"min_frp"`
	} `json:"firms"`
}

func loadServerConfig(path string) (serverConfig, error) {
	var c serverConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	if c.DataDir == "" {
		c.DataDir = filepath.Join(filepath.Dir(path), "data")
	}
	if c.Satellite.Layer == "" {
		c.Satellite.Layer = "nowcoast_global_longwave"
	}
	return c, nil
}

func main() {
	refresh := flag.Bool("refresh", false, "fetch all feeds and write the data dir, then exit (cron mode)")
	useFCGI := flag.Bool("fcgi", false, "serve over FastCGI (DreamHost dispatch.fcgi mode)")
	httpAddr := flag.String("http", "", "serve over plain HTTP at this address (dev mode), e.g. :8080")
	cfgPath := flag.String("config", "", "path to geisd config (keep outside the web root; required)")
	flag.Parse()
	log.SetFlags(log.LstdFlags)

	if *cfgPath == "" {
		log.Fatalf("-config is required (e.g. -config /home/dh_r7wsei/geis/geisd-config.json)")
	}
	cfg, err := loadServerConfig(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *refresh {
		if err := doRefresh(cfg); err != nil {
			log.Fatalf("refresh: %v", err)
		}
		return
	}

	handler := normalizePath(newMux(cfg.DataDir))
	switch {
	case *httpAddr != "":
		log.Printf("geisd serving HTTP on %s (data: %s)", *httpAddr, cfg.DataDir)
		log.Fatal(http.ListenAndServe(*httpAddr, handler))
	default:
		// FastCGI for both -fcgi and the no-flag case (Apache execs dispatch.fcgi
		// with no arguments). fcgi.Serve(nil, ...) speaks the protocol over the
		// socket the web server passes on stdin.
		_ = useFCGI
		if err := fcgi.Serve(nil, handler); err != nil {
			log.Fatalf("fcgi: %v", err)
		}
	}
}

// normalizePath strips a "/dispatch.fcgi" script-name prefix that DreamHost's
// FastCGI rewrite leaves on the request path, so the mux matches clean routes
// like /payload.json regardless of the .htaccess rewrite style.
func normalizePath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only strip the prefix at a path boundary, so "/dispatch.fcgiX" is not
		// mangled into "X".
		if r.URL.Path == "/dispatch.fcgi" {
			r.URL.Path = "/"
		} else if strings.HasPrefix(r.URL.Path, "/dispatch.fcgi/") {
			r.URL.Path = r.URL.Path[len("/dispatch.fcgi"):]
		}
		next.ServeHTTP(w, r)
	})
}

// doRefresh fetches every enabled feed and writes the served artifacts.
func doRefresh(cfg serverConfig) error {
	store := cache.New(cfg.DataDir)
	now := time.Now()

	var prev feeds.Payload
	if data, _, ok := store.Read("payload.json"); ok {
		if err := json.Unmarshal(data, &prev); err != nil {
			log.Printf("warn: corrupt cached payload, ignoring: %v", err)
		}
	}
	p := feeds.Payload{GeneratedAt: now}

	fetch := func(name string, fn func(context.Context) error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := fn(ctx); err != nil {
			log.Printf("%s: FAILED (%v) — kept last-good", name, err)
			return
		}
		log.Printf("%s: ok", name)
	}

	if cfg.Overlays.Storms {
		fetch("storms", func(ctx context.Context) error {
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
		fetch("quakes", func(ctx context.Context) error {
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
		fetch("aurora", func(ctx context.Context) error {
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
		fetch("iss", func(ctx context.Context) error {
			iss, err := feeds.FetchISS(ctx)
			if err != nil {
				p.ISS = prev.ISS
				return err
			}
			p.ISS = iss
			return nil
		})
	}
	if cfg.Overlays.Ships {
		fetch("ships", func(ctx context.Context) error {
			ships, err := feeds.FetchShips(ctx, feeds.AISHubConfig{
				Username: cfg.AISHub.Username,
				LatMin:   cfg.AISHub.LatMin, LatMax: cfg.AISHub.LatMax,
				LonMin: cfg.AISHub.LonMin, LonMax: cfg.AISHub.LonMax,
			})
			if err != nil {
				p.Ships = prev.Ships
				return err
			}
			p.Ships = ships
			return nil
		})
	}
	if cfg.Overlays.Fires {
		fetch("fires", func(ctx context.Context) error {
			fires, err := feeds.FetchFires(ctx, feeds.FIRMSConfig{
				MapKey: cfg.FIRMS.MapKey, Source: cfg.FIRMS.Source,
				Area: cfg.FIRMS.Area, DayRange: cfg.FIRMS.DayRange,
				MinFRP: cfg.FIRMS.MinFRP,
			})
			if err != nil {
				p.Fires = prev.Fires
				return err
			}
			p.Fires = fires
			return nil
		})
	}
	var satWriteErr error
	if cfg.Overlays.Satellite {
		fetch("satellite", func(ctx context.Context) error {
			img, err := feeds.FetchSatellite(ctx, cfg.Satellite.Layer, now)
			if err != nil {
				return err // fetch failure: keep last-good satellite.png
			}
			if err := store.Write("satellite.png", img, now); err != nil {
				satWriteErr = err // disk error: surface it below
				return err
			}
			return nil
		})
	}

	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := store.Write("payload.json", payloadJSON, now); err != nil {
		return err
	}
	shipsJSON, _ := json.Marshal(map[string]any{"ships": p.Ships, "generated_at": now})
	if err := store.Write("ships.json", shipsJSON, now); err != nil {
		return err
	}
	log.Printf("refresh done: storms=%d quakes=%d aurora=%d iss=%v ships=%d fires=%d",
		len(p.Storms), len(p.Quakes), len(p.Aurora), p.ISS != nil, len(p.Ships), len(p.Fires))
	if satWriteErr != nil {
		return fmt.Errorf("satellite write failed: %w", satWriteErr)
	}
	return nil
}

// newMux serves the static artifacts from dataDir. ServeFile provides correct
// Last-Modified/ETag handling for cheap revalidation.
func newMux(dataDir string) *http.ServeMux {
	mux := http.NewServeMux()
	serveFile := func(name, contentType string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			path := filepath.Join(dataDir, name)
			if _, err := os.Stat(path); err != nil {
				http.Error(w, "not generated yet", http.StatusServiceUnavailable)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=120")
			http.ServeFile(w, r, path)
		}
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/payload.json", serveFile("payload.json", "application/json"))
	mux.HandleFunc("/ships.json", serveFile("ships.json", "application/json"))
	mux.HandleFunc("/satellite.png", serveFile("satellite.png", "image/png"))
	mux.HandleFunc("/", indexHandler(dataDir))
	return mux
}

func indexHandler(dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		var p feeds.Payload
		if data, err := os.ReadFile(filepath.Join(dataDir, "payload.json")); err == nil {
			if err := json.Unmarshal(data, &p); err != nil {
				log.Printf("warn: failed to parse payload.json: %v", err)
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!doctype html><meta charset=utf-8>
<title>GEIS companion</title>
<style>body{font:16px/1.5 system-ui;margin:2rem;max-width:48rem}img{max-width:100%%;border:1px solid #ccc}</style>
<h1>GEIS — Global Event Information System</h1>
<p>Aggregated event feed for the GEIS desktop wallpaper.
Generated: %s</p>
<ul>
<li>Tropical cyclones: %d</li>
<li>Earthquakes: %d</li>
<li>Aurora cells: %d</li>
<li>ISS: %v</li>
<li>Ships: %d</li>
<li>Fires: %d</li>
</ul>
<p>Endpoints: <a href="/payload.json">/payload.json</a>,
<a href="/ships.json">/ships.json</a>,
<a href="/satellite.png">/satellite.png</a>,
<a href="/healthz">/healthz</a></p>
<img src="/satellite.png" alt="latest cloud imagery">
`, p.GeneratedAt.Format(time.RFC1123), len(p.Storms), len(p.Quakes), len(p.Aurora), p.ISS != nil, len(p.Ships), len(p.Fires))
	}
}
