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

	return encodeImage(dc.Image(), format, quality)
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
	val, ok := data[v.Name]
	if !ok || val == nil {
		return v.DefaultValue
	}
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
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
