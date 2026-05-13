package services

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/iZcy/imposizcy/internal/models"
	"github.com/sirupsen/logrus"
)

func writeWhitePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.White)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func newTestRenderer(t *testing.T, uploadDir string) *NativeRenderer {
	t.Helper()
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	logger.SetLevel(logrus.WarnLevel)
	reg, err := NewFontRegistry(nil, uploadDir, logger)
	if err != nil {
		t.Fatalf("font registry: %v", err)
	}
	return NewNativeRenderer(logger, reg, NewBarcodeService(), uploadDir)
}

func TestRender_TextOnly(t *testing.T) {
	dir := t.TempDir()
	bg := filepath.Join(dir, "bg.png")
	writeWhitePNG(t, bg, 200, 100)

	tpl := &models.PrintTemplate{
		BackgroundImage: "bg.png",
		Variables: []models.TemplateVariable{{
			Name:     "title",
			Type:     models.VariableTypeText,
			Position: &models.FieldPosition{X: 10, Y: 30, Width: 180, Height: 40, FontSize: 20, Alignment: "center", Color: "#000000"},
		}},
	}
	r := newTestRenderer(t, dir)
	out, err := r.Render(context.Background(), tpl, map[string]interface{}{"title": "HELLO"}, "png", 90)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Sample a band of pixels across the horizontal center; at least one should be dark.
	cy := img.Bounds().Dy() / 2
	dark := 0
	for x := 0; x < img.Bounds().Dx(); x++ {
		r, g, b, _ := img.At(x, cy).RGBA()
		if r < 0x4000 && g < 0x4000 && b < 0x4000 {
			dark++
		}
	}
	if dark == 0 {
		t.Fatalf("expected dark pixels along text band, found none")
	}
}

func TestRender_QR(t *testing.T) {
	dir := t.TempDir()
	bg := filepath.Join(dir, "bg.png")
	writeWhitePNG(t, bg, 250, 250)

	tpl := &models.PrintTemplate{
		BackgroundImage: "bg.png",
		Variables: []models.TemplateVariable{{
			Name:          "qr",
			Type:          models.VariableTypeBarcode,
			BarcodeFormat: models.BarcodeFormatQR,
			Position:      &models.FieldPosition{X: 25, Y: 25, Width: 200, Height: 200},
		}},
	}
	r := newTestRenderer(t, dir)
	out, err := r.Render(context.Background(), tpl, map[string]interface{}{"qr": "https://example.com"}, "png", 90)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// QR has finder squares in three corners. Verify a non-trivial number of dark pixels.
	dark := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 0x4000 && g < 0x4000 && b < 0x4000 {
				dark++
			}
		}
	}
	if dark < 200 {
		t.Fatalf("expected substantial dark pixels for QR, got %d", dark)
	}
}

func TestRender_FontFallback(t *testing.T) {
	dir := t.TempDir()
	bg := filepath.Join(dir, "bg.png")
	writeWhitePNG(t, bg, 200, 80)

	tpl := &models.PrintTemplate{
		BackgroundImage: "bg.png",
		Variables: []models.TemplateVariable{{
			Name:     "name",
			Type:     models.VariableTypeText,
			Position: &models.FieldPosition{X: 10, Y: 20, Width: 180, Height: 40, FontSize: 16},
			// FontFamily intentionally unknown — must fall back to embedded Go font.
			FontFamily: "NonexistentFamily",
		}},
	}
	r := newTestRenderer(t, dir)
	if _, err := r.Render(context.Background(), tpl, map[string]interface{}{"name": "Test"}, "png", 90); err != nil {
		t.Fatalf("expected fallback render to succeed, got: %v", err)
	}
}

func TestExpandDataRow_LengthMismatch(t *testing.T) {
	vars := []models.TemplateVariable{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	if _, err := ExpandDataRow([]interface{}{"x", "y"}, vars); err == nil {
		t.Fatalf("expected length-mismatch error")
	}
}

func TestExpandDataRow_OrderIndex(t *testing.T) {
	vars := []models.TemplateVariable{
		{Name: "first", OrderIndex: 2},
		{Name: "second", OrderIndex: 0},
		{Name: "third", OrderIndex: 1},
	}
	got, err := ExpandDataRow([]interface{}{"A", "B", "C"}, vars)
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	if got["second"] != "A" || got["third"] != "B" || got["first"] != "C" {
		t.Fatalf("OrderIndex ordering not honored: %#v", got)
	}
}
