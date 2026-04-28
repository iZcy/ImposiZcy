package services

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
	"github.com/sirupsen/logrus"
)

type ValidatorService struct {
	logger *logrus.Logger
}

func NewValidatorService(logger *logrus.Logger) *ValidatorService {
	return &ValidatorService{logger: logger}
}

func (s *ValidatorService) ValidateData(schemaStr string, data map[string]interface{}) (bool, []string, error) {
	schemaLoader := gojsonschema.NewStringLoader(schemaStr)

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return false, []string{err.Error()}, err
	}
	documentLoader := gojsonschema.NewBytesLoader(dataBytes)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, []string{err.Error()}, err
	}

	if result.Valid() {
		return true, nil, nil
	}

	var errors []string
	for _, desc := range result.Errors() {
		errors = append(errors, desc.String())
	}
	return false, errors, nil
}
