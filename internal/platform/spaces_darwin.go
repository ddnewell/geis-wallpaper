//go:build darwin

package platform

import (
	"bytes"
	"os"
	"path/filepath"

	"howett.net/plist"
)

// wallpaperStoreRel is the per-user wallpaper store WallpaperAgent reads.
const wallpaperStoreRel = "Library/Application Support/com.apple.wallpaper/Store/Index.plist"

// RegisterAllSpaces makes every desktop Space use the image at imagePath.
//
// macOS stores wallpaper per Space in Index.plist. A freshly-created Space is
// "Linked" (inherits the global default, which has no custom image), so a
// killall reload shows nothing for it. This copies a Space that already points
// at imagePath ("template") onto every Space that doesn't, so all desktops
// share the one stable file. It only rewrites the store when something actually
// changed (so steady-state runs are read-only), and never touches the global
// slots (forcing those corrupts the wallpaper state).
//
// Requires a template to already exist — created once interactively by
// `geis --register` (osascript). Returns whether the store was modified.
func RegisterAllSpaces(imagePath string) (changed bool, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	storePath := filepath.Join(home, wallpaperStoreRel)
	raw, err := os.ReadFile(storePath)
	if err != nil {
		return false, err
	}
	var root map[string]any
	if _, err := plist.Unmarshal(raw, &root); err != nil {
		return false, err
	}

	spaces, ok := root["Spaces"].(map[string]any)
	if !ok || len(spaces) == 0 {
		return false, nil
	}

	// Find a template Space whose Default already points at imagePath.
	var tmplDefault, tmplDisplays any
	for _, sv := range spaces {
		s, ok := sv.(map[string]any)
		if !ok {
			continue
		}
		if def, ok := s["Default"].(map[string]any); ok && slotImagePath(def) == imagePath {
			tmplDefault = deepCopy(def)
			tmplDisplays = deepCopy(s["Displays"])
			break
		}
	}
	if tmplDefault == nil {
		return false, nil // no template yet; `geis --register` creates one
	}

	for _, sv := range spaces {
		s, ok := sv.(map[string]any)
		if !ok {
			continue
		}
		def, _ := s["Default"].(map[string]any)
		if def != nil && slotImagePath(def) == imagePath {
			continue // already ours
		}
		s["Default"] = deepCopy(tmplDefault)
		if tmplDisplays != nil {
			s["Displays"] = deepCopy(tmplDisplays)
		}
		changed = true
	}
	if !changed {
		return false, nil
	}

	out, err := plist.Marshal(root, plist.BinaryFormat)
	if err != nil {
		return false, err
	}
	tmp := storePath + ".geistmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, storePath); err != nil {
		return false, err
	}
	return true, nil
}

// slotImagePath returns the image path a Default/Displays slot points at, or "".
func slotImagePath(slot map[string]any) string {
	desktop, ok := slot["Desktop"].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := desktop["Content"].(map[string]any)
	if !ok {
		return ""
	}
	choices, ok := content["Choices"].([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	blob, ok := choice["Configuration"].([]byte)
	if !ok || len(blob) == 0 {
		return ""
	}
	var cfg struct {
		URL struct {
			Relative string `plist:"relative"`
		} `plist:"url"`
	}
	if _, err := plist.Unmarshal(blob, &cfg); err != nil {
		return ""
	}
	rel := cfg.URL.Relative
	const pfx = "file://"
	if !bytes.HasPrefix([]byte(rel), []byte(pfx)) {
		return ""
	}
	return urlDecodePath(rel[len(pfx):])
}

// urlDecodePath turns a file URL path (with %20 etc.) back into a filesystem path.
func urlDecodePath(p string) string {
	out := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		if p[i] == '%' && i+2 < len(p) {
			if h := unhex(p[i+1])<<4 | unhex(p[i+2]); h >= 0 {
				out = append(out, byte(h))
				i += 2
				continue
			}
		}
		out = append(out, p[i])
	}
	return string(out)
}

func unhex(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	}
	return -1 << 8
}

// deepCopy clones the generic plist value graph so converted Spaces don't share
// references with the template.
func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, val := range x {
			m[k] = deepCopy(val)
		}
		return m
	case []any:
		s := make([]any, len(x))
		for i, val := range x {
			s[i] = deepCopy(val)
		}
		return s
	case []byte:
		b := make([]byte, len(x))
		copy(b, x)
		return b
	default:
		return x
	}
}
