package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chai2010/webp"
	"github.com/fogleman/gg"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/makiuchi-d/gozxing"
	gzcode128 "github.com/makiuchi-d/gozxing/oned"
	gzqr "github.com/makiuchi-d/gozxing/qrcode"
	"github.com/sirupsen/logrus"
)

// NativeRenderer composites text and barcodes directly onto a PNG background using gg + freetype.
// It is selected for templates whose RenderEngine == "native" (or whose BackgroundImage is set
// when RenderEngine is unspecified — lazy default).
type NativeRenderer struct {
	logger    *logrus.Logger
	fonts     *FontRegistry
	barcode   *BarcodeService
	uploadDir string
}

func NewNativeRenderer(logger *logrus.Logger, fonts *FontRegistry, barcode *BarcodeService, uploadDir string) *NativeRenderer {
	return &NativeRenderer{logger: logger, fonts: fonts, barcode: barcode, uploadDir: uploadDir}
}

// Render produces image bytes in the requested format ("png"|"jpeg"|"webp").
// Quality applies to jpeg/webp (1-100; default 90).
func (r *NativeRenderer) Render(ctx context.Context, t *models.PrintTemplate, data map[string]interface{}, format string, quality int) ([]byte, error) {
	if t.BackgroundImage == "" {
		return nil, fmt.Errorf("native renderer requires a background_image")
	}
	bg, err := r.loadBackground(t.BackgroundImage)
	if err != nil {
		return nil, fmt.Errorf("load background: %w", err)
	}
	width := bg.Bounds().Dx()
	height := bg.Bounds().Dy()

	dc := gg.NewContext(width, height)
	dc.DrawImage(bg, 0, 0)

	vars := orderedVariables(t.Variables)
	for _, v := range vars {
		// Shapes are static decoration and don't need a data value.
		if v.Type == models.VariableTypeShape {
			if err := r.drawShape(dc, v); err != nil {
				r.logger.WithError(err).WithField("variable", v.Name).Warn("shape draw failed")
			}
			continue
		}
		raw := resolveValue(v, data)
		if raw == "" {
			continue
		}
		switch v.Type {
		case models.VariableTypeBarcode:
			if err := r.drawBarcode(dc, v, raw); err != nil {
				r.logger.WithError(err).WithField("variable", v.Name).Warn("barcode draw failed")
			}
		case models.VariableTypeImage:
			if err := r.drawImage(dc, v, raw); err != nil {
				r.logger.WithError(err).WithField("variable", v.Name).Warn("image draw failed")
			}
		default: // text
			if err := r.drawText(ctx, dc, t, v, raw); err != nil {
				r.logger.WithError(err).WithField("variable", v.Name).Warn("text draw failed")
			}
		}
	}

	if err := r.verifyBarcodes(dc.Image(), vars, data); err != nil {
		return nil, err
	}

	return encodeImage(dc.Image(), format, quality)
}

// verifyBarcodes re-scans every barcode / QR variable from the freshly rendered canvas at its
// position bbox and asserts the decoded payload matches the source value. Guarantees that the
// emitted image is actually machine-readable and carries the right data — protects against silent
// encoder/format/size regressions.
func (r *NativeRenderer) verifyBarcodes(img image.Image, vars []models.TemplateVariable, data map[string]interface{}) error {
	type cropper interface {
		SubImage(image.Rectangle) image.Image
	}
	sub, ok := img.(cropper)
	if !ok {
		return nil
	}
	for _, v := range vars {
		if v.Type != models.VariableTypeBarcode || v.Position == nil {
			continue
		}
		want := resolveValue(v, data)
		if want == "" {
			continue
		}
		rect := image.Rect(int(v.Position.X), int(v.Position.Y),
			int(v.Position.X+v.Position.Width), int(v.Position.Y+v.Position.Height))
		region := sub.SubImage(rect)
		bmp, err := gozxing.NewBinaryBitmapFromImage(region)
		if err != nil {
			return fmt.Errorf("verify %q: bitmap: %w", v.Name, err)
		}
		var reader gozxing.Reader
		if v.BarcodeFormat == models.BarcodeFormatQR {
			reader = gzqr.NewQRCodeReader()
		} else {
			reader = gzcode128.NewCode128Reader()
		}
		result, err := reader.Decode(bmp, nil)
		if err != nil {
			return fmt.Errorf("verify %q (%s): cannot re-scan rendered barcode: %w", v.Name, v.BarcodeFormat, err)
		}
		got := result.GetText()
		if got != want {
			return fmt.Errorf("verify %q (%s): rendered payload %q != expected %q", v.Name, v.BarcodeFormat, got, want)
		}
		r.logger.WithFields(logrus.Fields{
			"variable": v.Name,
			"format":   v.BarcodeFormat,
			"payload":  got,
		}).Debug("barcode round-trip verified")
	}
	return nil
}

func (r *NativeRenderer) drawText(ctx context.Context, dc *gg.Context, t *models.PrintTemplate, v models.TemplateVariable, value string) error {
	pos := v.Position
	if pos == nil {
		return fmt.Errorf("text variable %q has no position", v.Name)
	}
	size := pos.FontSize
	if size <= 0 {
		size = v.FontSize
	}
	if size <= 0 {
		size = 14
	}
	bold := pos.Bold || strings.EqualFold(v.FontWeight, "bold")

	family := v.FontFamily
	if family == "" {
		family = t.DefaultFont
	}
	face := r.fonts.Face(ctx, family, size, bold)
	dc.SetFontFace(face)

	hex := pos.Color
	if hex == "" {
		hex = v.FontColor
	}
	if hex == "" {
		hex = "#000000"
	}
	dc.SetHexColor(hex)

	x := pos.X
	y := pos.Y
	w := pos.Width
	h := pos.Height
	if w <= 0 {
		w = float64(dc.Width()) - x
	}
	if h <= 0 {
		// Estimate one-line height if no bbox supplied.
		h = size * 1.4
	}

	align := strings.ToLower(pos.Alignment)
	if align == "" {
		align = strings.ToLower(v.TextAlign)
	}
	var ax float64
	switch align {
	case "center":
		ax = 0.5
	case "right":
		ax = 1.0
	default:
		ax = 0.0
	}

	// Word-wrap to width, then vertical-center the stack inside the bbox.
	lines := dc.WordWrap(value, w)
	if len(lines) == 0 {
		lines = []string{value}
	}
	lineH := size * 1.2
	stackH := lineH * float64(len(lines))
	startY := y + (h-stackH)/2 + size // baseline of first line; +size shifts from top to baseline

	for i, line := range lines {
		lx := x + ax*w
		ly := startY + float64(i)*lineH
		dc.DrawStringAnchored(line, lx, ly, ax, 0)
	}
	return nil
}

func (r *NativeRenderer) drawBarcode(dc *gg.Context, v models.TemplateVariable, value string) error {
	pos := v.Position
	if pos == nil {
		return fmt.Errorf("barcode variable %q has no position", v.Name)
	}
	w := int(pos.Width)
	h := int(pos.Height)
	if w <= 0 {
		w = 200
	}
	if h <= 0 {
		h = 80
	}
	format := string(v.BarcodeFormat)
	if format == "" {
		format = "code128"
	}
	img, err := r.barcode.RasterizePNG(value, format, w, h)
	if err != nil {
		return err
	}
	dc.DrawImage(img, int(pos.X), int(pos.Y))
	return nil
}

// drawShape renders a primitive (rect, line, ellipse) at v.Position with optional fill + stroke.
// For "line", (X,Y)..(X+Width,Y+Height) is the segment; Fill is ignored.
func (r *NativeRenderer) drawShape(dc *gg.Context, v models.TemplateVariable) error {
	pos := v.Position
	if pos == nil {
		return fmt.Errorf("shape %q has no position", v.Name)
	}
	kind := v.Shape
	if kind == "" {
		kind = models.ShapeRect
	}
	switch kind {
	case models.ShapeLine:
		dc.DrawLine(pos.X, pos.Y, pos.X+pos.Width, pos.Y+pos.Height)
	case models.ShapeEllipse:
		dc.DrawEllipse(pos.X+pos.Width/2, pos.Y+pos.Height/2, pos.Width/2, pos.Height/2)
	default: // rect
		dc.DrawRectangle(pos.X, pos.Y, pos.Width, pos.Height)
	}
	hasFill := v.Fill != "" && !strings.EqualFold(v.Fill, "none") && kind != models.ShapeLine
	hasStroke := v.Stroke != "" && !strings.EqualFold(v.Stroke, "none")
	if hasFill && hasStroke {
		dc.SetHexColor(v.Fill)
		dc.FillPreserve()
		sw := v.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		dc.SetLineWidth(sw)
		dc.SetHexColor(v.Stroke)
		dc.Stroke()
	} else if hasFill {
		dc.SetHexColor(v.Fill)
		dc.Fill()
	} else if hasStroke {
		sw := v.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		dc.SetLineWidth(sw)
		dc.SetHexColor(v.Stroke)
		dc.Stroke()
	} else {
		// nothing specified — clear the path so it doesn't get filled later
		dc.ClearPath()
	}
	return nil
}

func (r *NativeRenderer) drawImage(dc *gg.Context, v models.TemplateVariable, value string) error {
	pos := v.Position
	if pos == nil {
		return fmt.Errorf("image variable %q has no position", v.Name)
	}
	img, err := r.loadBackground(value)
	if err != nil {
		return err
	}
	w := int(pos.Width)
	h := int(pos.Height)
	if w > 0 && h > 0 && (img.Bounds().Dx() != w || img.Bounds().Dy() != h) {
		scaled := gg.NewContext(w, h)
		scaled.DrawImageAnchored(img, w/2, h/2, 0.5, 0.5)
		img = scaled.Image()
	}
	dc.DrawImage(img, int(pos.X), int(pos.Y))
	return nil
}

func (r *NativeRenderer) loadBackground(path string) (image.Image, error) {
	searchPaths := []string{path}
	if r.uploadDir != "" {
		searchPaths = append(searchPaths, filepath.Join(r.uploadDir, path))
	}
	searchPaths = append(searchPaths, filepath.Base(path))

	var full string
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			full = p
			break
		}
	}
	if full == "" {
		return nil, fmt.Errorf("background not found: %s", path)
	}
	return gg.LoadImage(full)
}

func orderedVariables(in []models.TemplateVariable) []models.TemplateVariable {
	out := make([]models.TemplateVariable, len(in))
	copy(out, in)
	// Stable sort by OrderIndex; zero values preserve declaration order.
	sort.SliceStable(out, func(i, j int) bool { return out[i].OrderIndex < out[j].OrderIndex })
	return out
}

func resolveValue(v models.TemplateVariable, data map[string]interface{}) string {
	keys := []string{v.Name}
	if v.SourceField != "" && v.SourceField != v.Name {
		keys = append([]string{v.SourceField}, keys...)
	}
	for _, k := range keys {
		val, ok := data[k]
		if !ok || val == nil {
			continue
		}
		if s, ok := val.(string); ok {
			if s != "" {
				return s
			}
			continue
		}
		return fmt.Sprintf("%v", val)
	}
	return v.DefaultValue
}

func encodeImage(img image.Image, format string, quality int) ([]byte, error) {
	if quality <= 0 || quality > 100 {
		quality = 90
	}
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	case "webp":
		if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: float32(quality)}); err != nil {
			return nil, err
		}
	default:
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// ExpandDataRow maps a positional []interface{} input to the keyed map[string]interface{}
// shape consumed by the renderer, using each variable's OrderIndex (fallback: declaration order).
// Returns an error if the row length doesn't match the variable count.
func ExpandDataRow(row []interface{}, variables []models.TemplateVariable) (map[string]interface{}, error) {
	if len(row) != len(variables) {
		return nil, fmt.Errorf("data_row length %d does not match %d variables", len(row), len(variables))
	}
	ordered := orderedVariables(variables)
	out := make(map[string]interface{}, len(row))
	for i, v := range ordered {
		out[v.Name] = row[i]
	}
	return out, nil
}
