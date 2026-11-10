package imaging

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func makeImage(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 90, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDownscalesLongestSide(t *testing.T) {
	out, err := Prepare(makeImage(t, 3000, 2000))
	if err != nil {
		t.Fatal(err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if format != "jpeg" {
		t.Errorf("format = %q, want jpeg", format)
	}
	if cfg.Width != MaxSide {
		t.Errorf("width = %d, want %d", cfg.Width, MaxSide)
	}
	if cfg.Height != 1066 {
		t.Errorf("height = %d, want 1066 (aspect preserved)", cfg.Height)
	}
}

func TestSmallImageStaysSameSize(t *testing.T) {
	out, err := Prepare(makeImage(t, 640, 480))
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 640 || cfg.Height != 480 {
		t.Fatalf("size = %dx%d", cfg.Width, cfg.Height)
	}
}

// The contract is fewer pixels, not fewer bytes: a synthetic gradient PNG can
// compress smaller than any JPEG. What matters is that vision never receives
// more than MaxSide.
func TestPixelCountIsBounded(t *testing.T) {
	out, err := Prepare(makeImage(t, 3000, 2000))
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width > MaxSide || cfg.Height > MaxSide {
		t.Fatalf("prepared %dx%d exceeds the %d px cap", cfg.Width, cfg.Height, MaxSide)
	}
	if cfg.Width*cfg.Height > 3000*2000 {
		t.Fatal("prepared image has more pixels than the input")
	}
}

func TestRejectsGarbage(t *testing.T) {
	if _, err := Prepare([]byte("not an image")); err == nil {
		t.Fatal("expected an error")
	}
}

func TestJPEGInputIsAccepted(t *testing.T) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 32, 32)), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Prepare(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}
