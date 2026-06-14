// Package config defines the GEIS configuration schema and loading logic.
//
// The schema intentionally covers all milestones (overlays, companion, paths)
// so later code slots in without breaking the file format. Only a subset is
// consumed by the offline MVP (M1): displays, darkness, solar, paths, output.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Screen is a single physical display target in pixels.
type Screen struct {
	Width   int  `json:"width"`
	Height  int  `json:"height"`
	Primary bool `json:"primary"`
}

// Displays controls how many wallpapers are produced and at what size.
type Displays struct {
	AutoDetect       bool     `json:"auto_detect"`
	Screens          []Screen `json:"screens"`
	RenderPerDisplay bool     `json:"render_per_display"`
}

// SolarConfig tunes the irradiance heatmap layer.
type SolarConfig struct {
	// Colormap selects the day-side heatmap gradient ("solar", "inferno", "grayscale").
	Colormap string `json:"colormap"`
	// DayAlpha is the maximum opacity of the day-side heatmap tint (0..1).
	DayAlpha float64 `json:"day_alpha"`
	// MaxGHI is the W/m^2 value mapped to the top of the colormap.
	MaxGHI float64 `json:"max_ghi"`
	// TwilightDeg is the solar-elevation span (degrees below horizon) over which
	// night darkening fades in. 18 = astronomical twilight.
	TwilightDeg float64 `json:"twilight_deg"`
}

// Overlays toggles each event layer.
type Overlays struct {
	Satellite   bool `json:"satellite"`
	Storms      bool `json:"storms"`
	Earthquakes bool `json:"earthquakes"`
	Aurora      bool `json:"aurora"`
	ISS         bool `json:"iss"`
	Ships       bool `json:"ships"`
	Fires       bool `json:"fires"`
}

// SatelliteConfig selects the cloud layer and its blend.
type SatelliteConfig struct {
	Layer string  `json:"layer"`
	Alpha float64 `json:"alpha"`
}

// Earthquakes filters the USGS feed.
type EarthquakesConfig struct {
	MinMagnitude float64 `json:"min_magnitude"`
}

// Companion points the client at the optional aggregation server.
type Companion struct {
	Enabled bool   `json:"enabled"`
	BaseURL string `json:"base_url"`
}

// Paths overrides the default on-disk locations. Empty values use defaults.
type Paths struct {
	AssetDir   string `json:"asset_dir"`
	CacheDir   string `json:"cache_dir"`
	OutputDir  string `json:"output_dir"`
	ClocksFile string `json:"clocks_file"`
}

// Output controls the unique-frame ring buffer.
type Output struct {
	KeepLast int `json:"keep_last"`
}

// Config is the full GEIS configuration.
type Config struct {
	Displays            Displays          `json:"displays"`
	Darkness            float64           `json:"darkness"`
	Solar               SolarConfig       `json:"solar"`
	SatelliteAboveSolar bool              `json:"satellite_above_solar"`
	Overlays            Overlays          `json:"overlays"`
	Satellite           SatelliteConfig   `json:"satellite"`
	Earthquakes         EarthquakesConfig `json:"earthquakes"`
	Companion           Companion         `json:"companion"`
	Paths               Paths             `json:"paths"`
	Output              Output            `json:"output"`
}

// Default returns a config with sensible values for a fresh install.
func Default() Config {
	c := Config{}
	c.Displays.AutoDetect = true
	c.Darkness = 0.667
	c.Solar = SolarConfig{Colormap: "solar", DayAlpha: 0.5, MaxGHI: 1100, TwilightDeg: 18}
	c.SatelliteAboveSolar = false
	c.Overlays = Overlays{Satellite: true, Storms: true, Earthquakes: true, Aurora: true, ISS: true, Ships: false, Fires: false}
	c.Satellite = SatelliteConfig{Layer: "nowcoast_global_longwave", Alpha: 0.35}
	c.Earthquakes = EarthquakesConfig{MinMagnitude: 4.0}
	c.Companion = Companion{Enabled: false, BaseURL: ""}
	c.Output = Output{KeepLast: 3}
	return c
}

// Load reads a config file over the defaults. A missing path returns defaults.
func Load(path string) (Config, error) {
	c := Default()
	if path == "" {
		return c, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.Output.KeepLast < 1 {
		c.Output.KeepLast = 3
	}
	if c.Solar.MaxGHI <= 0 {
		c.Solar.MaxGHI = 1100
	}
	if c.Solar.TwilightDeg <= 0 {
		c.Solar.TwilightDeg = 18
	}
	return c, nil
}

// supportRoot is the per-user GEIS data root.
func supportRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Application Support", "GEIS")
}

// AssetDir resolves the directory holding img/, ico/, json/. Defaults to the
// directory of the running binary so an in-repo `./geis` works during dev,
// falling back to the install root.
func (c Config) AssetDir() string {
	if c.Paths.AssetDir != "" {
		return expand(c.Paths.AssetDir)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		// In-repo build: binary sits beside img/ and json/.
		if _, err := os.Stat(filepath.Join(dir, "img")); err == nil {
			return dir
		}
	}
	if wd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(wd, "img")); err == nil {
			return wd
		}
	}
	return supportRoot()
}

// CacheDir resolves where fetched overlays are cached.
func (c Config) CacheDir() string {
	if c.Paths.CacheDir != "" {
		return expand(c.Paths.CacheDir)
	}
	return filepath.Join(supportRoot(), "cache")
}

// OutputDir resolves where unique wallpaper frames are written. It must be a
// non-purgeable location (Application Support, not /tmp or ~/Library/Caches).
func (c Config) OutputDir() string {
	if c.Paths.OutputDir != "" {
		return expand(c.Paths.OutputDir)
	}
	return filepath.Join(supportRoot(), "frames")
}

// ClocksFile resolves the world-clock definition file.
func (c Config) ClocksFile() string {
	if c.Paths.ClocksFile != "" {
		return expand(c.Paths.ClocksFile)
	}
	return filepath.Join(c.AssetDir(), "json", "clocks.json")
}

// expand resolves a leading ~ to the user home directory.
func expand(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}
