//go:build !darwin

package platform

import "fmt"

var errUnsupported = fmt.Errorf("wallpaper control is only supported on macOS")

// SetWallpaper is unsupported off macOS.
func SetWallpaper(path string) error { return errUnsupported }

// SetWallpaperForScreen is unsupported off macOS.
func SetWallpaperForScreen(idx int, path string) error { return errUnsupported }

// ScreenCount is unsupported off macOS.
func ScreenCount() int { return 0 }

// ScreenSize is unsupported off macOS.
func ScreenSize(idx int) (w, h int, err error) { return 0, 0, errUnsupported }

// MainScreenSize is unsupported off macOS.
func MainScreenSize() (w, h int, err error) { return 0, 0, errUnsupported }

// SetWallpaperAllDesktops is unsupported off macOS.
func SetWallpaperAllDesktops(path string) error { return errUnsupported }

// RefreshSpaces is a no-op off macOS.
func RefreshSpaces() error { return nil }
