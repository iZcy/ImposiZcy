package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/iZcy/imposizcy/internal/models"
	"github.com/sirupsen/logrus"
)

type RendererService struct {
	logger         *logrus.Logger
	barcodeService *BarcodeService
	uploadDir      string
}

func NewRendererService(logger *logrus.Logger, barcodeService *BarcodeService, uploadDir string) *RendererService {
	return &RendererService{
		logger:         logger,
		barcodeService: barcodeService,
		uploadDir:      uploadDir,
	}
}

func (s *RendererService) Render(templateStr string, data map[string]interface{}) (string, error) {
	return s.renderWithData(templateStr, data)
}

func (s *RendererService) renderWithData(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("print").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// variableRegex matches {{variable_name}} (without dot prefix)
var variableRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

func (s *RendererService) RenderHTML(htmlTemplate string, css string, variables []models.TemplateVariable, data map[string]interface{}) (string, error) {
	processedData := make(map[string]interface{})
	for k, v := range data {
		processedData[k] = v
	}

	for _, variable := range variables {
		val, ok := data[variable.Name]
		if !ok {
			continue
		}
		strVal, ok := val.(string)
		if !ok {
			strVal = fmt.Sprintf("%v", val)
		}

		switch variable.Type {
		case models.VariableTypeBarcode:
			format := string(variable.BarcodeFormat)
			if format == "" {
				format = string(models.BarcodeFormatCode128)
			}
			bw, bh := 200, 80
			if format == "qr" {
				bw, bh = 200, 200
			}
			dataURI, err := s.barcodeService.GenerateBarcode(strVal, format, bw, bh)
			if err != nil {
				s.logger.WithError(err).WithField("variable", variable.Name).Error("Failed to generate barcode")
				processedData[variable.Name] = ""
				continue
			}
			processedData[variable.Name] = fmt.Sprintf(`<img src="%s" alt="barcode" style="max-width:100%%;height:auto;" />`, dataURI)
		default:
			processedData[variable.Name] = strVal
		}
	}

	// Convert {{variable}} to {{.variable}} for Go template compatibility
	processedHTML := variableRegex.ReplaceAllString(htmlTemplate, "{{.$1}}")

	tmpl, err := template.New("print").Parse(processedHTML)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, processedData); err != nil {
		return "", err
	}

	html := buf.String()
	if css != "" {
		html = fmt.Sprintf("<style>%s</style>%s", css, html)
	}

	return html, nil
}

// RenderPositioned generates an HTML page with a background image and absolutely positioned variable overlays.
// This is the DocuSign-style positioning: each variable has x, y, width, height relative to the template canvas.
func (s *RendererService) RenderPositioned(tmpl *models.PrintTemplate, data map[string]interface{}) (string, error) {
	if tmpl.BackgroundImage == "" {
		return "", fmt.Errorf("template has no background image")
	}

	// Resolve the background image to a file:// URI or data URI
	bgURI, err := s.resolveImageURI(tmpl.BackgroundImage)
	if err != nil {
		return "", fmt.Errorf("failed to resolve background image: %w", err)
	}

	// Build positioned overlays for each variable
	var overlays strings.Builder
	for _, v := range tmpl.Variables {
		value, ok := data[v.Name]
		if !ok {
			if v.DefaultValue != "" {
				value = v.DefaultValue
			} else {
				continue
			}
		}
		strVal, ok := value.(string)
		if !ok {
			strVal = fmt.Sprintf("%v", value)
		}

		overlay := s.buildVariableOverlay(v, strVal)
		overlays.WriteString(overlay)
	}

	// Calculate actual pixel dimensions
	widthPx := tmpl.Width
	heightPx := tmpl.Height
	if tmpl.DimensionUnit == models.DimensionUnitMM && tmpl.DPI > 0 {
		widthPx = tmpl.Width * float64(tmpl.DPI) / 25.4
		heightPx = tmpl.Height * float64(tmpl.DPI) / 25.4
	} else if tmpl.DimensionUnit == models.DimensionUnitCM && tmpl.DPI > 0 {
		widthPx = tmpl.Width * float64(tmpl.DPI) / 2.54
		heightPx = tmpl.Height * float64(tmpl.DPI) / 2.54
	} else if tmpl.DimensionUnit == models.DimensionUnitInch && tmpl.DPI > 0 {
		widthPx = tmpl.Width * float64(tmpl.DPI)
		heightPx = tmpl.Height * float64(tmpl.DPI)
	}
	if widthPx == 0 {
		widthPx = 800
	}
	if heightPx == 0 {
		heightPx = 400
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { width: %[1]gpx; height: %[2]gpx; overflow: hidden; position: relative; }
  .bg-image {
    position: absolute; top: 0; left: 0;
    width: %[1]gpx; height: %[2]gpx;
    z-index: 0;
  }
  .bg-image img {
    width: %[1]gpx; height: %[2]gpx;
    object-fit: fill;
  }
  .field-overlay {
    position: absolute;
    overflow: hidden;
    display: flex;
    z-index: 1;
  }
  .field-overlay.text {
    word-break: break-word;
    white-space: pre-wrap;
  }
  .field-overlay.barcode img {
    width: 100%%;
    height: 100%%;
    object-fit: contain;
  }
</style>
</head>
<body>
<div class="bg-image"><img src="%[3]s" /></div>
%[4]s
</body>
</html>`, widthPx, heightPx, bgURI, overlays.String())

	return html, nil
}

// buildVariableOverlay creates the absolute-positioned HTML div for a single variable
func (s *RendererService) buildVariableOverlay(v models.TemplateVariable, value string) string {
	pos := v.Position
	if pos == nil {
		// No position defined — skip overlay
		return ""
	}

	x := pos.X
	y := pos.Y
	w := pos.Width
	h := pos.Height
	if w == 0 {
		w = 200
	}
	if h == 0 {
		h = 40
	}

	fontSize := pos.FontSize
	if fontSize == 0 {
		fontSize = 16
	}

	align := pos.Alignment
	if align == "" {
		align = "left"
	}

	// Map alignment to flex justify
	flexAlign := "flex-start"
	switch align {
	case "center":
		flexAlign = "center"
	case "right":
		flexAlign = "flex-end"
	}

	content := ""
	cssClass := "text"

	switch v.Type {
	case models.VariableTypeBarcode:
		cssClass = "barcode"
		format := string(v.BarcodeFormat)
		if format == "" {
			format = string(models.BarcodeFormatCode128)
		}
		// Use the overlay dimensions for barcode size
		dataURI, err := s.barcodeService.GenerateBarcode(value, format, int(w), int(h))
		if err != nil {
			s.logger.WithError(err).WithField("variable", v.Name).Error("Failed to generate barcode for overlay")
			content = html.EscapeString(value)
		} else {
			content = fmt.Sprintf(`<img src="%s" alt="%s" />`, dataURI, html.EscapeString(v.Name))
		}

	case models.VariableTypeImage:
		// Resolve image path to data URI
		imgURI, err := s.resolveImageURI(value)
		if err != nil {
			content = html.EscapeString(value)
		} else {
			content = fmt.Sprintf(`<img src="%s" alt="%s" style="width:100%%;height:100%%;object-fit:contain;" />`, imgURI, html.EscapeString(v.Name))
		}

	default: // text
		content = html.EscapeString(value)
	}

	return fmt.Sprintf(
		`<div class="field-overlay %s" style="left:%gpx;top:%gpx;width:%gpx;height:%gpx;font-size:%dpx;align-items:center;justify-content:%s;color:%s;">%s</div>`+"\n",
		cssClass, x, y, w, h, fontSize, flexAlign, pos.Color, content,
	)
}

// resolveImageURI converts a relative upload path or file path to a usable URI for HTML
func (s *RendererService) resolveImageURI(imagePath string) (string, error) {
	// If it's already a data URI, return as-is
	if strings.HasPrefix(imagePath, "data:") {
		return imagePath, nil
	}
	// If it's already an http(s) URL, return as-is
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		return imagePath, nil
	}

	// Try to read the file and convert to data URI
	searchPaths := []string{imagePath}
	if s.uploadDir != "" {
		searchPaths = append(searchPaths, filepath.Join(s.uploadDir, imagePath))
	}
	searchPaths = append(searchPaths, filepath.Base(imagePath))

	var fullPath string
	var data []byte
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			fullPath = p
			break
		}
	}

	if fullPath == "" {
		s.logger.WithField("path", imagePath).Warn("Could not resolve image file")
		return imagePath, nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return imagePath, nil
	}

	mimeType := "image/png"
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".bmp":
		mimeType = "image/bmp"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

// ApplyFieldMapping transforms flat incoming data using the template's field mapping.
// If field_mapping is defined, maps source_key → target variable name.
// If no mapping is defined, passes data through unchanged.
func ApplyFieldMapping(data map[string]interface{}, mapping []models.FieldMapping) map[string]interface{} {
	if len(mapping) == 0 {
		return data
	}

	result := make(map[string]interface{})
	for _, m := range mapping {
		val, ok := data[m.SourceField]
		if !ok {
			if m.DefaultValue != "" {
				result[m.TargetVariable] = m.DefaultValue
			}
			continue
		}
		result[m.TargetVariable] = val
	}
	return result
}
