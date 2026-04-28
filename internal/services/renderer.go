package services

import (
	"bytes"
	"text/template"

	"github.com/sirupsen/logrus"
)

type RendererService struct {
	logger *logrus.Logger
}

func NewRendererService(logger *logrus.Logger) *RendererService {
	return &RendererService{logger: logger}
}

func (s *RendererService) Render(templateStr string, data map[string]interface{}) (string, error) {
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

func (s *RendererService) RenderHTML(htmlTemplate string, data map[string]interface{}) (string, error) {
	return s.Render(htmlTemplate, data)
}
