// Package imaging prepares an upload for the vision model.
//
// ELA always runs on the original bytes. The model gets a downscaled JPEG:
// vision cost and latency scale with pixel count, and a 1600 px copy carries
// the same text and layout as a 4096 px screenshot.
package imaging

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png" // register decoder

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register decoder
)

// MaxSide is the longest edge sent to the model.
const MaxSide = 1600

// Quality keeps text legible while shrinking the base64 payload.
const Quality = 85

// ErrDecode means the bytes are not a decodable image.
var ErrDecode = errors.New("image decode failed")

// Prepare returns a JPEG suitable for the vision model. Images already small
// enough are re-encoded at Quality so the payload stays predictable.
func Prepare(encoded []byte) ([]byte, error) {
	src, _, err := image.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, ErrDecode
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil, ErrDecode
	}

	targetW, targetH := w, h
	if w > MaxSide || h > MaxSide {
		if w >= h {
			targetW = MaxSide
			targetH = max(1, h*MaxSide/w)
		} else {
			targetH = MaxSide
			targetW = max(1, w*MaxSide/h)
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: Quality}); err != nil {
		return nil, ErrDecode
	}
	return buf.Bytes(), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
