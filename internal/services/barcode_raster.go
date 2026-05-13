package services

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
)

// RasterizePNG produces a raster image for the supplied barcode content / format.
// Output is an *image.RGBA sized exactly to width x height, white background, black bars/modules.
func (s *BarcodeService) RasterizePNG(content, format string, width, height int) (image.Image, error) {
	format = strings.ToLower(format)
	if width <= 0 {
		width = 200
	}
	if height <= 0 {
		height = 80
	}
	switch format {
	case "qr":
		return s.rasterQR(content, width, height)
	case "code128", "":
		bars, err := encodeCode128(content)
		if err != nil {
			return nil, err
		}
		return s.barsToImage(bars, width, height, content), nil
	case "ean13":
		bars := buildEAN13Bars(content)
		return s.barsToImage(bars, width, height, ean13Digits(content)), nil
	case "code39":
		if content == "" {
			return nil, fmt.Errorf("Code 39 content cannot be empty")
		}
		bars := buildCode39Bars(content)
		return s.barsToImage(bars, width, height, content), nil
	default:
		return nil, fmt.Errorf("unsupported barcode format %q", format)
	}
}

// barsToImage mirrors barsToSVG: each slice entry is the width of a black bar (in units),
// followed by an implicit 1-unit white separator. A 20px footer holds the human-readable label.
func (s *BarcodeService) barsToImage(bars []int, width, height int, label string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	if len(bars) == 0 {
		drawCenteredLabel(img, label)
		return img
	}

	totalUnits := 0
	for _, b := range bars {
		totalUnits += b
		if b > 0 {
			totalUnits++ // 1-unit space after each bar
		}
	}
	if totalUnits == 0 {
		drawCenteredLabel(img, label)
		return img
	}

	unit := float64(width) / float64(totalUnits)
	if unit < 1 {
		unit = 1
	}
	barRect := image.Rect(0, 0, 1, height-20)
	if height-20 <= 0 {
		barRect = image.Rect(0, 0, 1, height)
	}
	black := &image.Uniform{C: color.Black}

	x := 0.0
	for _, bar := range bars {
		if bar > 0 {
			w := unit * float64(bar)
			rx0 := int(x + 0.5)
			rx1 := int(x + w + 0.5)
			if rx1 > width {
				rx1 = width
			}
			if rx1 > rx0 {
				r := image.Rect(rx0, barRect.Min.Y, rx1, barRect.Max.Y)
				draw.Draw(img, r, black, image.Point{}, draw.Src)
			}
			x += w
		}
		x += unit
	}

	if label != "" && height-20 > 0 {
		drawFooterLabel(img, label)
	}
	return img
}

func (s *BarcodeService) rasterQR(content string, width, height int) (image.Image, error) {
	modules, size, err := generateQR(content)
	if err != nil {
		return nil, err
	}
	// Mirror generateQRCode: pad with 4-module quiet zone on each side.
	totalModules := size + 8
	side := width
	if height < side {
		side = height
	}
	cellSize := side / totalModules
	if cellSize < 1 {
		cellSize = 1
	}
	canvasSide := cellSize * totalModules
	// Output canvas is the requested rectangle, centered QR inside.
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	offX := (width-canvasSide)/2 + cellSize*4
	offY := (height-canvasSide)/2 + cellSize*4
	black := &image.Uniform{C: color.Black}
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if modules[row*size+col] {
				r := image.Rect(offX+col*cellSize, offY+row*cellSize, offX+(col+1)*cellSize, offY+(row+1)*cellSize)
				draw.Draw(img, r, black, image.Point{}, draw.Src)
			}
		}
	}
	return img, nil
}

// drawCenteredLabel renders text at the image center using a small mono font.
func drawCenteredLabel(img *image.RGBA, label string) {
	if label == "" {
		return
	}
	dc := gg.NewContextForRGBA(img)
	dc.SetFontFace(monoFace(12))
	dc.SetColor(color.Black)
	w, _ := dc.MeasureString(label)
	dc.DrawString(label, float64(img.Bounds().Dx())/2-w/2, float64(img.Bounds().Dy())/2+6)
}

func drawFooterLabel(img *image.RGBA, label string) {
	dc := gg.NewContextForRGBA(img)
	dc.SetFontFace(monoFace(12))
	dc.SetColor(color.Black)
	w, _ := dc.MeasureString(label)
	dc.DrawString(label, float64(img.Bounds().Dx())/2-w/2, float64(img.Bounds().Dy())-5)
}

var (
	monoOnce sync.Once
	monoFont *truetype.Font
)

func monoFace(size float64) font.Face {
	monoOnce.Do(func() {
		if tt, err := truetype.Parse(gomono.TTF); err == nil {
			monoFont = tt
		}
	})
	if monoFont == nil {
		return nil
	}
	return truetype.NewFace(monoFont, &truetype.Options{Size: size, DPI: 72})
}

// ean13 + code39 helpers re-create the bar slices used by the existing SVG generators
// so the raster path produces visually identical output.

func ean13Digits(content string) string {
	digits := ""
	for _, ch := range content {
		if ch >= '0' && ch <= '9' {
			digits += string(ch)
		}
	}
	if len(digits) < 13 {
		digits = strings.Repeat("0", 13-len(digits)) + digits
	} else if len(digits) > 13 {
		digits = digits[:13]
	}
	return digits
}

func buildEAN13Bars(content string) []int {
	digits := ean13Digits(content)
	bars := []int{}
	bars = append(bars, 1, 0, 1)
	bs := BarcodeService{}
	for _, d := range digits[:6] {
		bars = append(bars, bs.eanPattern(int(d-'0'))...)
		bars = append(bars, 0)
	}
	bars = append(bars, 0, 1, 0, 1, 0)
	for _, d := range digits[6:] {
		bars = append(bars, bs.eanPattern(int(d-'0'))...)
		bars = append(bars, 0)
	}
	bars = append(bars, 1, 0, 1)
	return bars
}

func buildCode39Bars(content string) []int {
	code39Map := map[rune][]int{
		'0': {1, 0, 0, 1, 0, 1, 1, 0, 1}, '1': {1, 1, 0, 0, 1, 0, 1, 0, 1}, '2': {1, 0, 0, 1, 1, 0, 1, 0, 1},
		'3': {1, 1, 0, 1, 1, 0, 1, 0, 0}, '4': {1, 0, 0, 1, 0, 1, 0, 1, 1}, '5': {1, 1, 0, 1, 0, 1, 0, 0, 1},
		'6': {1, 0, 0, 1, 1, 0, 0, 1, 1}, '7': {1, 1, 0, 0, 1, 0, 1, 1, 0}, '8': {1, 0, 0, 1, 0, 1, 1, 1, 0},
		'9': {1, 1, 0, 1, 0, 0, 1, 1, 0}, '-': {1, 0, 0, 1, 0, 0, 1, 1, 1}, '.': {1, 1, 0, 0, 1, 0, 0, 1, 1},
		' ': {1, 1, 0, 0, 1, 1, 0, 0, 1}, '$': {1, 0, 0, 1, 0, 0, 1, 0, 1}, '/': {1, 0, 0, 1, 0, 1, 0, 0, 1},
		'+': {1, 0, 0, 0, 1, 0, 1, 0, 1}, '%': {1, 0, 1, 0, 0, 1, 0, 0, 1},
	}
	bars := []int{1, 0, 0, 1, 0, 1, 1, 0, 1, 0}
	for _, ch := range strings.ToUpper(content) {
		if pattern, ok := code39Map[ch]; ok {
			bars = append(bars, pattern...)
			bars = append(bars, 0)
		}
	}
	bars = append(bars, 1, 0, 0, 1, 0, 1, 1, 0, 1)
	return bars
}
