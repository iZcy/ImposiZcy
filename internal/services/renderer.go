package services

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"github.com/iZcy/imposizcy/internal/models"
	"github.com/sirupsen/logrus"
)

type RendererService struct {
	logger         *logrus.Logger
	barcodeService *BarcodeService
}

func NewRendererService(logger *logrus.Logger, barcodeService *BarcodeService) *RendererService {
	return &RendererService{
		logger:         logger,
		barcodeService: barcodeService,
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
			dataURI, err := s.barcodeService.GenerateBarcode(strVal, format, 200, 80)
			if err != nil {
				s.logger.WithError(err).WithField("variable", variable.Name).Error("Failed to generate barcode")
				processedData[variable.Name] = ""
				continue
			}
			processedData[variable.Name] = dataURI
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
