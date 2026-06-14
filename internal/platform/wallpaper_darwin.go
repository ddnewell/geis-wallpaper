//go:build darwin

// Package platform contains the macOS-specific integration: setting the desktop
// wallpaper via the official NSWorkspace API, reading the main display size,
// advisory locking, and the unique-frame ring buffer.
//
// NSWorkspace.setDesktopImageURL is used (not osascript) because it is the
// official API, requires no TCC/AppleEvent prompt, and is not affected by the
// macOS Tahoe regression where AppleScript only updates the active Space after
// the first run. Note: macOS exposes no public API to set ALL Spaces, so this
// sets the current Space of each display every cycle.
package platform

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include <stdlib.h>
#include "wallpaper_darwin.h"
*/
import "C"
import (
	"fmt"
	"os/exec"
	"unsafe"
)

// SetWallpaperAllDesktops sets the same image on every display's current Space
// via System Events. Empirically this persists to the wallpaper store more
// reliably than NSWorkspace on macOS Tahoe, so the path survives a subsequent
// WallpaperAgent reload — which is what lets RefreshSpaces propagate the update
// to every Space.
func SetWallpaperAllDesktops(path string) error {
	script := fmt.Sprintf(`tell application "System Events" to tell every desktop to set picture to POSIX file %q`, path)
	if out, err := exec.Command("/usr/bin/osascript", "-e", script).CombinedOutput(); err != nil {
		return fmt.Errorf("osascript set wallpaper: %v: %s", err, out)
	}
	return nil
}

// RefreshSpaces forces WallpaperAgent to reload the desktop image from disk for
// every Space (it relaunches automatically within ~1s). With a stable image
// path (already persisted by SetWallpaperAllDesktops), this both bypasses the
// macOS same-path image cache and propagates the in-place content update beyond
// the current Space to all desktops.
func RefreshSpaces() error {
	// killall exits non-zero if no process matched, which is harmless here.
	_ = exec.Command("/usr/bin/killall", "WallpaperAgent").Run()
	return nil
}

// SetWallpaper sets the desktop image for every screen's current Space.
func SetWallpaper(path string) error {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	rc := int(C.geis_set_wallpaper_all(cs))
	if rc < 0 {
		return fmt.Errorf("no screens detected")
	}
	if rc > 0 {
		return fmt.Errorf("failed to set wallpaper on %d screen(s)", rc)
	}
	return nil
}

// SetWallpaperForScreen sets the desktop image for one screen by index.
func SetWallpaperForScreen(idx int, path string) error {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	switch int(C.geis_set_wallpaper_for_screen(C.int(idx), cs)) {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("screen index %d out of range", idx)
	default:
		return fmt.Errorf("failed to set wallpaper on screen %d", idx)
	}
}

// ScreenCount returns the number of attached displays.
func ScreenCount() int { return int(C.geis_screen_count()) }

// ScreenSize returns the backing-store (pixel) size of screen idx (idx<0 = main).
func ScreenSize(idx int) (w, h int, err error) {
	var cw, ch C.int
	if rc := int(C.geis_screen_size(C.int(idx), &cw, &ch)); rc != 0 {
		return 0, 0, fmt.Errorf("could not read size of screen %d", idx)
	}
	return int(cw), int(ch), nil
}

// MainScreenSize returns the main display's backing-store (pixel) dimensions.
func MainScreenSize() (w, h int, err error) { return ScreenSize(-1) }
