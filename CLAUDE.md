# geis-wallpaper

Go rewrite (formerly Python 2) of a macOS desktop-wallpaper generator: an
equirectangular world map shaded by real-time solar irradiance, with cached
event overlays and an optional DreamHost FastCGI companion. The original Python
is archived under `old/` (still tracked) — ignore it for new work.

## Layout
- `cmd/geis` — renderer (cgo/NSWorkspace; runs under launchd; never hits network)
- `cmd/geis-fetch` — fetcher (feeds → local cache; deliberately decoupled from render)
- `cmd/geisd` — companion server (`-refresh` cron mode, `-fcgi`/`-http` serve modes)
- `internal/{config,geo,solar,render,cache,feeds,fonts,platform}`

## Build / test / run
- `go build ./... && go test ./... && go vet ./...`; `gofmt -l . | grep -v '^old/'` must be empty
- Renderer needs cgo (darwin + Xcode CLT). Companion is cgo-free — cross-compile:
  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o geisd ./cmd/geisd`
- `./geis --dry-run` renders without setting the wallpaper; output at
  `~/Library/Application Support/GEIS/frames/wallpaper.png` (view it to verify)
- Companion locally: `./geisd -refresh -config <cfg>` then `./geisd -http :8099 -config <cfg>`
- launchd: agents `at.newell.geis.{changewallpaper,fetch}`; reload via
  `launchctl bootout`/`bootstrap gui/$(id -u)/<label>`, force-run `launchctl kickstart -k`,
  logs in `~/Library/Logs/GEIS/`

## macOS wallpaper gotchas (hard-won — read before touching platform/)
- Map is equirectangular EPSG:4326; lon/lat→pixel is one linear map. Go image origin
  is top-left, so `y=(90-lat)/180*H` (matplotlib used bottom-left — easy to mirror).
- Wallpaper is stored PER SPACE in `~/Library/Application Support/com.apple.wallpaper/Store/Index.plist`.
  To cover all Spaces, write the image into each Space's slot + `killall WallpaperAgent`.
  NEVER force the global `AllSpacesAndDisplays`/`SystemDefault` slots — it corrupts the
  wallpaper state. Back up Index.plist before editing it.
- NSWorkspace sets only the *live* image; only osascript/System Events writes the persisted
  store. osascript→System Events HANGS under launchd (no Automation/TCC grant) — keep
  osascript out of per-cycle launchd jobs; `geis --register` does the one-time osascript
  step interactively at install.
- Use a STABLE wallpaper path + `killall WallpaperAgent` to refresh content; unique
  filenames make Spaces diverge.

## Data sources / deploy
- Keyless feeds run client-side (NHC CurrentStorms.json, NASA GIBS / NOAA nowCoast WMS
  EPSG:4326, USGS, NOAA SWPC, wheretheiss.at); keyed feeds (AISHub, NASA FIRMS) run
  server-side only. Verify endpoints are actually live before trusting them, and bump
  `http.Transport.TLSHandshakeTimeout` above Go's 10s default (some hosts are slow).
- Deploy target: DreamHost shared host (SSH alias `geis-weldedanvil-com`), docroot
  `/home/dh_r7wsei/geis/prod`, binary/config/data/secrets in `/home/dh_r7wsei/geis/`.
  See `deploy/dreamhost/README.md`. Never commit API keys.
