geis-wallpaper
==============

**Global Event Information System** — renders a live world map as your macOS
desktop wallpaper: a Natural Earth base map shaded by the **present-time solar
irradiance** (a clear-sky W/m² heatmap with a physically-placed day/night
terminator), overlaid with **clouds, tropical cyclones, earthquakes, the aurora
oval, the ISS, ships, and wildfires**, plus a row of **world clocks**.

It is a from-scratch Go rewrite of the original 2012 Python/matplotlib/cartopy
project. The wallpaper update runs **fully offline** and at minimal cost: solar
and clocks are pure local math, and event overlays come from a local cache that
degrades gracefully when there's no network.

Verified on **macOS 26 "Tahoe"** (Apple Silicon), Go 1.26.

How it works
------------

Two tiny per-user launchd agents, deliberately decoupled so a slow network can
never block the wallpaper:

- **`geis`** (renderer, every ~1 min) — *never touches the network*. Reads the
  base map, `clocks.json`, and the local cache; computes the solar layer for
  *now*; composites everything; writes the PNG to a stable path; sets it via the
  official `NSWorkspace.setDesktopImageURL` API (no permission prompt); then
  reloads `WallpaperAgent` so the update lands on **every Space**, not just the
  active one.
- **`geis-fetch`** (fetcher, every ~15 min) — pulls event data into the cache
  with atomic writes and hard timeouts. Each feed falls back to its last-good
  value, so a transient failure dims nothing.

Everything is rendered on one 2560×1280 equirectangular (EPSG:4326) grid, so
`lon/lat → pixel` is a single linear map — no projection library needed. The
cloud imagery (NASA GIBS / NOAA nowCoast) is served in the same projection at
the same size, so layers composite pixel-for-pixel.

Optionally, a **Go FastCGI companion** (`geisd`, see `deploy/dreamhost/`)
aggregates all feeds server-side — including the keyed ones (AIS ships, NASA
FIRMS fires) whose secrets stay off every client — and serves one clean
`/payload.json` + `/satellite.png`.

Install (macOS)
---------------

```sh
deploy/install.sh
```

This builds `geis` (+ `geis-fetch`), copies the binaries and assets to
`~/Library/Application Support/GEIS`, writes `config.json` (from
`json/config.example.json` if absent), installs the two LaunchAgents, and runs
once. `NSWorkspace` needs no permission prompt, so the first run just works.

Uninstall with `deploy/uninstall.sh`.

Configuration
-------------

`~/Library/Application Support/GEIS/config.json` (see `json/config.example.json`):

| Key | Meaning |
|---|---|
| `displays.auto_detect` | size output to the main display (default true) |
| `displays.render_per_display` | render a per-aspect image for each monitor |
| `darkness` | max night-side darkening (0–1, default 0.667) |
| `solar.colormap` | `solar` / `inferno` / `grayscale` heatmap |
| `solar.day_alpha` | opacity of the day-side irradiance heatmap |
| `solar.max_ghi` | W/m² mapped to the top of the colormap |
| `overlays.*` | toggle satellite, storms, earthquakes, aurora, iss, ships, fires |
| `satellite.layer` | `nowcoast_global_longwave` (IR) or `gibs_truecolor` |
| `companion.enabled` / `base_url` | use the companion server for data |
| `output.keep_last` | unique-frame ring depth (default 3) |

Data sources (all free)
-----------------------

Keyless, fetched by the client directly or via the companion: tropical cyclones
(NOAA NHC `CurrentStorms.json`), clouds (NOAA nowCoast / NASA GIBS WMS),
earthquakes (USGS GeoJSON), aurora oval (NOAA SWPC OVATION), ISS
(wheretheiss.at). Key-gated, server-side only: ships (AISHub), wildfires (NASA
FIRMS).

Offline behavior
----------------

The renderer always produces a complete wallpaper with **zero** network:
solar + clocks are local; cached overlays are drawn with an "as of" note and
dropped only if older than 24 h. With no cache at all (fresh install offline),
overlays are simply omitted.

Known limitations
-----------------

- macOS has no public API to set *all Spaces*, and stores wallpaper **per
  Space**. GEIS handles this in three parts: (1) `geis --register` (run once at
  install; may ask for Automation access — the only time osascript is used)
  points the current Space at a **stable image path**; (2) each cycle the
  renderer writes that path into every desktop Space's slot in the wallpaper
  store (`com.apple.wallpaper/Store/Index.plist`) — no permissions needed, and
  it only rewrites when a Space is missing it, so new desktops are picked up
  automatically; (3) it overwrites the image in place and reloads
  `WallpaperAgent`, which re-reads it across every Space. Side effect: a brief
  wallpaper reload (~1 s) each minute — raise the renderer's `StartInterval` if
  you find it distracting.
- AIS ships require a free AISHub membership; wildfires require a free FIRMS key.
  Both are configured on the companion server, not the client.

Development
-----------

```sh
go test ./...
go build -o geis ./cmd/geis && ./geis --dry-run   # render without setting the desktop
```

The original Python implementation is preserved under `old/` for reference.
License: MIT.
