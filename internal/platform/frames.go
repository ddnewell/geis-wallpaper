package platform

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// WriteStable writes img to dir/name, atomically (temp + rename) overwriting in
// place. Using a STABLE path (not a unique one per cycle) keeps every Space
// pointed at the same file, so a WallpaperAgent reload refreshes them all
// together. The macOS same-path image cache is defeated by RefreshSpaces
// (killall WallpaperAgent), which forces a fresh read from disk — so unique
// filenames (which made Spaces diverge) are no longer needed.
func WriteStable(dir, name string, img image.Image) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(f, img); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", err
	}
	return path, nil
}

// CleanupFrames removes wallpaper PNGs in dir not present in keep — both legacy
// timestamped frames from the older unique-filename scheme and per-display
// files no longer in use.
func CleanupFrames(dir string, keep map[string]bool) {
	for _, pat := range []string{"geis-*.png", "wallpaper*.png"} {
		matches, _ := filepath.Glob(filepath.Join(dir, pat))
		for _, p := range matches {
			if !keep[p] {
				_ = os.Remove(p)
			}
		}
	}
}
