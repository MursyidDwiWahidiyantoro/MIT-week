package forensics

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"testing"
)

// flatImage is a uniform grey field; noisyPatch draws random pixels in a
// rectangle, which is exactly the "pasted region" ELA is meant to expose.
func flatImage(w, h int, patch image.Rectangle) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}
	rng := rand.New(rand.NewSource(42))
	for y := patch.Min.Y; y < patch.Max.Y; y++ {
		for x := patch.Min.X; x < patch.Max.X; x++ {
			img.Set(x, y, color.RGBA{
				uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255,
			})
		}
	}
	return img
}

func encodeJPEG(t *testing.T, img image.Image, q int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: q}); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

func encodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}
	return buf.Bytes()
}

// ELA compares the image against a q=90 re-encode of itself. If the input was
// already saved at q=90 the re-encode is a no-op and every difference is ~0,
// so the test input must carry a different compression history — which real
// screenshots and re-saved photos always do.
const testInputQuality = 75

func TestPastedPatchScoresHigherThanFlat(t *testing.T) {
	flat := encodeJPEG(t, flatImage(256, 256, image.Rect(0, 0, 0, 0)), testInputQuality)
	patched := encodeJPEG(t, flatImage(256, 256, image.Rect(64, 64, 192, 192)), testInputQuality)

	flatRes, err := Analyze(flat)
	if err != nil {
		t.Fatalf("flat: %v", err)
	}
	patchRes, err := Analyze(patched)
	if err != nil {
		t.Fatalf("patched: %v", err)
	}

	if patchRes.Score <= flatRes.Score {
		t.Fatalf("patched score %d should exceed flat score %d", patchRes.Score, flatRes.Score)
	}
	if peak(patchRes) <= peak(flatRes) {
		t.Fatalf("patched peak %.3f should exceed flat peak %.3f", peak(patchRes), peak(flatRes))
	}
}

// peak returns the hottest grid cell.
func peak(r Result) float64 {
	var max float64
	for y := 0; y < GridSize; y++ {
		for x := 0; x < GridSize; x++ {
			if r.Grid[y][x] > max {
				max = r.Grid[y][x]
			}
		}
	}
	return max
}

// An image saved at exactly the analysis quality has no measurable difference.
// This is a property of the technique, not a bug: it is why the UI must show a
// reliability note rather than implying ELA always sees something.
func TestSameQualityInputIsUnmeasurable(t *testing.T) {
	img := flatImage(256, 256, image.Rect(64, 64, 192, 192))
	res, err := Analyze(encodeJPEG(t, img, JPEGQuality))
	if err != nil {
		t.Fatal(err)
	}
	if res.Score != 0 {
		t.Errorf("score = %d, want 0 when the input is already at q=%d", res.Score, JPEGQuality)
	}
}

func TestReliabilityByFormat(t *testing.T) {
	img := flatImage(128, 128, image.Rect(0, 0, 0, 0))

	jpegRes, err := Analyze(encodeJPEG(t, img, 90))
	if err != nil {
		t.Fatal(err)
	}
	if jpegRes.Reliability != 1.0 {
		t.Errorf("jpeg reliability = %v, want 1.0", jpegRes.Reliability)
	}

	pngRes, err := Analyze(encodePNG(t, img))
	if err != nil {
		t.Fatal(err)
	}
	if pngRes.Reliability != 0.5 {
		t.Errorf("png reliability = %v, want 0.5", pngRes.Reliability)
	}
	if pngRes.Note == "" {
		t.Error("png result should carry a reliability note")
	}
}

func TestTinyImageHasZeroReliability(t *testing.T) {
	res, err := Analyze(encodePNG(t, flatImage(32, 32, image.Rect(0, 0, 0, 0))))
	if err != nil {
		t.Fatal(err)
	}
	if res.Reliability != 0 {
		t.Errorf("reliability = %v, want 0 for an image under 64 px", res.Reliability)
	}
	if res.Note == "" {
		t.Error("expected a note explaining the image is too small")
	}
}

func TestOversizeRejected(t *testing.T) {
	big := encodePNG(t, flatImage(MaxSide+1, 8, image.Rect(0, 0, 0, 0)))
	if _, err := Analyze(big); err != ErrTooLarge {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
}

func TestScoreWithinRange(t *testing.T) {
	res, err := Analyze(encodeJPEG(t, flatImage(200, 200, image.Rect(10, 10, 190, 190)), 50))
	if err != nil {
		t.Fatal(err)
	}
	if res.Score < 0 || res.Score > 100 {
		t.Fatalf("score = %d, out of range", res.Score)
	}
}

func TestHotspotsAreNormalizedAndInsideBounds(t *testing.T) {
	res, err := Analyze(encodeJPEG(t, flatImage(256, 256, image.Rect(64, 64, 192, 192)), 90))
	if err != nil {
		t.Fatal(err)
	}
	hotspots := res.Hotspots(1.5)
	if len(hotspots) == 0 {
		t.Fatal("expected at least one hotspot on a patched image")
	}
	for _, box := range hotspots {
		ymin, xmin, ymax, xmax := box[0], box[1], box[2], box[3]
		if ymin < 0 || xmin < 0 || ymax > 1000 || xmax > 1000 {
			t.Errorf("box %v outside the 0-1000 range", box)
		}
		if ymax <= ymin || xmax <= xmin {
			t.Errorf("box %v is degenerate", box)
		}
	}
}

func TestGridIsSixteenBySixteen(t *testing.T) {
	res, err := Analyze(encodeJPEG(t, flatImage(160, 160, image.Rect(0, 0, 0, 0)), 90))
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Grid) != GridSize || len(res.Grid[0]) != GridSize {
		t.Fatalf("grid = %dx%d, want %dx%d", len(res.Grid), len(res.Grid[0]), GridSize, GridSize)
	}
}
