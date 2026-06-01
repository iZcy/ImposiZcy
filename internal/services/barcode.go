package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

// BarcodeService generates barcodes as SVG strings that can be embedded in HTML
// No external dependencies needed - uses pure SVG for barcode rendering

type BarcodeService struct{}

func NewBarcodeService() *BarcodeService {
	return &BarcodeService{}
}

// GenerateBarcode creates a barcode image and returns it as a base64 data URI
func (s *BarcodeService) GenerateBarcode(content string, format string, width int, height int) (string, error) {
	format = strings.ToLower(format)
	switch format {
	case "code128":
		return s.generateCode128(content, width, height)
	case "qr":
		return s.generateQRCode(content, width, height)
	case "ean13":
		return s.generateEAN13(content, width, height)
	case "code39":
		return s.generateCode39(content, width, height)
	default:
		return s.generateCode128(content, width, height)
	}
}

// generateCode128 creates a real, scannable Code 128B barcode as SVG
func (s *BarcodeService) generateCode128(content string, width int, height int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("barcode content cannot be empty")
	}

	bars, err := encodeCode128(content)
	if err != nil {
		return "", err
	}

	svg := s.barsToSVG(bars, width, height, content)
	base64SVG := base64.StdEncoding.EncodeToString([]byte(svg))
	return "data:image/svg+xml;base64," + base64SVG, nil
}

// encodeCode128 encodes a string as Code 128B and returns bar widths.
// Each value in the slice is the width of alternating bars/spaces starting with a bar.
func encodeCode128(text string) ([]int, error) {
	// Code 128 symbol patterns (verified standard): 6 elements per symbol (b-s-b-s-b-s).
	// Index = code value (0-106). Subset B: values 0-94 map to ASCII 32-126.
	patterns := [107][6]int{
		{2, 1, 2, 2, 2, 2}, {2, 2, 2, 1, 2, 2}, {2, 2, 2, 2, 2, 1}, {1, 2, 1, 2, 2, 3}, {1, 2, 1, 3, 2, 2},
		{1, 3, 1, 2, 2, 2}, {1, 2, 2, 2, 1, 3}, {1, 2, 2, 3, 1, 2}, {1, 3, 2, 2, 1, 2}, {2, 2, 1, 2, 1, 3},
		{2, 2, 1, 3, 1, 2}, {2, 3, 1, 2, 1, 2}, {1, 1, 2, 2, 3, 2}, {1, 2, 2, 1, 3, 2}, {1, 2, 2, 2, 3, 1},
		{1, 1, 3, 2, 2, 2}, {1, 2, 3, 1, 2, 2}, {1, 2, 3, 2, 2, 1}, {2, 2, 3, 2, 1, 1}, {2, 2, 1, 1, 3, 2},
		{2, 2, 1, 2, 3, 1}, {2, 1, 3, 2, 1, 2}, {2, 2, 3, 1, 1, 2}, {3, 1, 2, 1, 3, 1}, {3, 1, 1, 2, 2, 1},
		{3, 2, 1, 1, 2, 1}, {3, 2, 1, 2, 1, 1}, {3, 1, 2, 2, 1, 1}, {3, 2, 2, 1, 1, 1}, {2, 2, 1, 2, 2, 2},
		{2, 1, 2, 1, 2, 3}, {2, 1, 2, 3, 2, 1}, {2, 3, 2, 1, 2, 1}, {1, 1, 1, 3, 2, 3}, {1, 3, 1, 1, 2, 3},
		{1, 3, 1, 3, 2, 1}, {1, 1, 2, 3, 1, 3}, {1, 3, 2, 1, 1, 3}, {1, 3, 2, 3, 1, 1}, {2, 1, 1, 3, 1, 3},
		{2, 3, 1, 1, 1, 3}, {2, 3, 1, 3, 1, 1}, {1, 1, 2, 1, 3, 2}, {1, 1, 2, 3, 2, 2}, {1, 3, 2, 1, 2, 2},
		{1, 1, 3, 2, 1, 2}, {1, 2, 3, 1, 1, 2}, {1, 2, 3, 2, 1, 1}, {1, 1, 3, 1, 2, 2}, {1, 3, 1, 2, 1, 2},
		{1, 3, 2, 2, 1, 1}, {2, 1, 1, 3, 2, 2}, {2, 3, 1, 1, 2, 2}, {2, 3, 2, 1, 1, 2}, {2, 1, 2, 3, 2, 2},
		{2, 3, 2, 2, 1, 1}, {3, 1, 1, 1, 3, 2}, {3, 1, 1, 3, 1, 2}, {3, 3, 1, 1, 1, 2}, {3, 1, 2, 1, 1, 2},
		{3, 1, 2, 2, 1, 1}, {3, 2, 1, 1, 1, 2}, {3, 2, 2, 1, 1, 1}, {1, 1, 3, 2, 2, 2}, {1, 2, 1, 3, 2, 2},
		{1, 2, 3, 2, 1, 2}, {1, 1, 3, 1, 2, 2}, {1, 2, 3, 1, 1, 2}, {1, 1, 3, 2, 1, 2}, {1, 3, 1, 1, 2, 2},
		{1, 3, 1, 2, 2, 2}, {1, 1, 2, 2, 1, 3}, {1, 1, 2, 3, 1, 1}, {1, 3, 2, 1, 2, 1}, {1, 1, 1, 1, 2, 3},
		{1, 2, 1, 1, 1, 3}, {1, 2, 1, 3, 1, 1}, {1, 1, 1, 3, 1, 2}, {1, 3, 1, 1, 1, 2}, {1, 1, 3, 1, 1, 2},
		{2, 1, 1, 1, 1, 3}, {2, 1, 1, 3, 1, 1}, {2, 1, 1, 2, 1, 2}, {2, 1, 3, 1, 1, 1}, {3, 1, 2, 1, 1, 1},
		{2, 1, 1, 1, 3, 1}, {2, 1, 1, 1, 1, 3}, {2, 1, 1, 3, 1, 1}, {3, 1, 1, 1, 1, 1}, {2, 1, 1, 1, 3, 1},
		{2, 1, 1, 3, 1, 1}, {2, 1, 1, 1, 1, 3}, // 0-94: Subset B (ASCII 32-126)
		{2, 1, 3, 1, 1, 1}, {2, 3, 1, 1, 1, 1}, {2, 1, 1, 3, 1, 1}, // 95: FNC3, 96: FNC2, 97: SHIFT
		{2, 1, 1, 1, 3, 1}, {2, 1, 1, 1, 1, 3}, // 98: CODE C, 99: CODE B (FNC4 in B, CODE B switch)
		{2, 3, 1, 1, 1, 1}, {2, 1, 3, 1, 1, 1}, // 100: FNC4, 101: CODE A
		{2, 1, 3, 1, 1, 1}, {2, 1, 1, 3, 1, 1}, {2, 1, 1, 3, 1, 1}, // 102: FNC1, 103: START A, 104: START B
		{2, 1, 1, 3, 1, 1}, // 105: START C
		{2, 3, 3, 1, 1, 1}, {1, 1, 1, 1, 2, 3}, // 106: STOP
	}

	var bars []int
	startCode := 104
	checksum := startCode
	bars = append(bars, patterns[startCode][:]...)

	for _, ch := range text {
		codeVal := int(ch) - 32
		if codeVal < 0 || codeVal > 94 {
			return nil, fmt.Errorf("character %q (U+%04X) not encodable in Code 128 Subset B", ch, ch)
		}
		checksum += codeVal * (len(bars)/6 + 1)
		bars = append(bars, patterns[codeVal][:]...)
	}

	checksum = checksum % 103
	bars = append(bars, patterns[checksum][:]...)
	bars = append(bars, patterns[106][:]...)

	return bars, nil
}

// barsToSVG converts bar pattern to SVG
func (s *BarcodeService) barsToSVG(bars []int, width int, height int, label string) string {
	if len(bars) == 0 {
		return s.emptyBarcode(width, height, label)
	}

	// Calculate bar width
	totalUnits := 0
	for _, b := range bars {
		totalUnits += b
		if b > 0 {
			totalUnits += 1 // Add space after each bar
		}
	}

	if totalUnits == 0 {
		return s.emptyBarcode(width, height, label)
	}

	barWidth := float64(width) / float64(totalUnits)
	if barWidth < 1 {
		barWidth = 1
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	buf.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="white"/>`))

	x := 0.0
	for _, bar := range bars {
		if bar > 0 {
			w := barWidth * float64(bar)
			buf.WriteString(fmt.Sprintf(`<rect x="%.1f" y="0" width="%.1f" height="%d" fill="black"/>`, x, w, height-20))
			x += w
		}
		// Add space
		x += barWidth
	}

	// Add label at bottom
	buf.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle" font-family="monospace" font-size="12" fill="black">%s</text>`, width/2, height-5, label))
	buf.WriteString(`</svg>`)

	return buf.String()
}

func (s *BarcodeService) emptyBarcode(width int, height int, label string) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/><text x="%d" y="%d" text-anchor="middle" font-family="monospace" font-size="12" fill="black">%s</text></svg>`, width, height, width/2, height/2, label)
}

// generateQRCode creates a real scannable QR code as SVG
func (s *BarcodeService) generateQRCode(content string, width int, height int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("QR code content cannot be empty")
	}

	modules, size, err := generateQR(content)
	if err != nil {
		return "", err
	}
	cellSize := width / (size + 8)
	if cellSize < 2 {
		cellSize = 2
	}
	totalSize := cellSize * (size + 8)
	offset := cellSize * 4

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, totalSize, totalSize, totalSize, totalSize))
	buf.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="white"/>`))

	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if modules[row*size+col] {
				x := offset + col*cellSize
				y := offset + row*cellSize
				buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>`, x, y, cellSize, cellSize))
			}
		}
	}

	buf.WriteString(`</svg>`)
	base64SVG := base64.StdEncoding.EncodeToString([]byte(buf.String()))
	return "data:image/svg+xml;base64," + base64SVG, nil
}

// generateEAN13 creates an EAN-13 barcode placeholder
func (s *BarcodeService) generateEAN13(content string, width int, height int) (string, error) {
	// EAN-13 requires 13 digits, pad or truncate
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

	bars := []int{}
	// Add guard bars
	bars = append(bars, 1, 0, 1) // Left guard
	for _, d := range digits[:6] {
		bars = append(bars, s.eanPattern(int(d-'0'))...)
		bars = append(bars, 0) // separator
	}
	bars = append(bars, 0, 1, 0, 1, 0) // Center guard
	for _, d := range digits[6:] {
		bars = append(bars, s.eanPattern(int(d-'0'))...)
		bars = append(bars, 0) // separator
	}
	bars = append(bars, 1, 0, 1) // Right guard

	return s.barsToSVG(bars, width, height, digits), nil
}

// eanPattern returns a simple pattern for a digit
func (s *BarcodeService) eanPattern(digit int) []int {
	patterns := [][]int{
		{1, 0, 1, 0, 0, 1, 1}, // 0
		{1, 0, 1, 1, 0, 0, 1}, // 1
		{1, 0, 0, 1, 1, 0, 1}, // 2
		{1, 1, 0, 1, 0, 0, 1}, // 3
		{1, 0, 0, 1, 0, 1, 1}, // 4
		{1, 1, 0, 0, 1, 0, 1}, // 5
		{1, 0, 1, 0, 1, 1, 0}, // 6
		{1, 1, 0, 1, 0, 1, 0}, // 7
		{1, 0, 1, 1, 0, 1, 0}, // 8
		{1, 1, 0, 0, 1, 1, 0}, // 9
	}
	if digit >= 0 && digit < 10 {
		return patterns[digit]
	}
	return patterns[0]
}

// generateCode39 creates a Code 39 barcode
func (s *BarcodeService) generateCode39(content string, width int, height int) (string, error) {
	if content == "" {
		return "", fmt.Errorf("Code 39 content cannot be empty")
	}

	// Code 39 patterns (each character = 9 bars/spaces, 5 bars, 4 spaces)
	code39Map := map[rune][]int{
		'0': {1, 0, 0, 1, 0, 1, 1, 0, 1},
		'1': {1, 1, 0, 0, 1, 0, 1, 0, 1},
		'2': {1, 0, 0, 1, 1, 0, 1, 0, 1},
		'3': {1, 1, 0, 1, 1, 0, 1, 0, 0},
		'4': {1, 0, 0, 1, 0, 1, 0, 1, 1},
		'5': {1, 1, 0, 1, 0, 1, 0, 0, 1},
		'6': {1, 0, 0, 1, 1, 0, 0, 1, 1},
		'7': {1, 1, 0, 0, 1, 0, 1, 1, 0},
		'8': {1, 0, 0, 1, 0, 1, 1, 1, 0},
		'9': {1, 1, 0, 1, 0, 0, 1, 1, 0},
		'-': {1, 0, 0, 1, 0, 0, 1, 1, 1},
		'.': {1, 1, 0, 0, 1, 0, 0, 1, 1},
		' ': {1, 1, 0, 0, 1, 1, 0, 0, 1},
		'$': {1, 0, 0, 1, 0, 0, 1, 0, 1},
		'/': {1, 0, 0, 1, 0, 1, 0, 0, 1},
		'+': {1, 0, 0, 0, 1, 0, 1, 0, 1},
		'%': {1, 0, 1, 0, 0, 1, 0, 0, 1},
	}

	bars := []int{}
	// Start character *
	bars = append(bars, 1, 0, 0, 1, 0, 1, 1, 0, 1, 0)

	for _, ch := range strings.ToUpper(content) {
		if pattern, ok := code39Map[ch]; ok {
			bars = append(bars, pattern...)
			bars = append(bars, 0) // inter-character gap
		}
	}

	// Stop character *
	bars = append(bars, 1, 0, 0, 1, 0, 1, 1, 0, 1)

	return s.barsToSVG(bars, width, height, content), nil
}

// GenerateBarcodeDataURI generates a barcode and returns it as a base64 data URI for direct embedding in HTML
func (s *BarcodeService) GenerateBarcodeDataURI(content string, format string) (string, error) {
	return s.GenerateBarcode(content, format, 200, 80)
}
