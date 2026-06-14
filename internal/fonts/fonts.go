// Package fonts provides cached opentype faces. It uses the Go font shipped
// with golang.org/x/image, so no external TTF file or go:embed is required and
// text rendering works fully offline.
package fonts

import (
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var (
	loadOnce    sync.Once
	regularFont *opentype.Font
	boldFont    *opentype.Font

	mu    sync.Mutex
	faces = map[key]font.Face{}
)

type key struct {
	bold bool
	size float64
}

func load() {
	regularFont, _ = opentype.Parse(goregular.TTF)
	boldFont, _ = opentype.Parse(gobold.TTF)
}

// Face returns a cached face at the given pixel size. DPI is fixed at 72 so
// Size is interpreted directly in pixels.
func Face(size float64, bold bool) font.Face {
	loadOnce.Do(load)
	mu.Lock()
	defer mu.Unlock()
	k := key{bold, size}
	if f, ok := faces[k]; ok {
		return f
	}
	src := regularFont
	if bold {
		src = boldFont
	}
	f, err := opentype.NewFace(src, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil
	}
	faces[k] = f
	return f
}
