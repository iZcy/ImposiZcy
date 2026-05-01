package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type KafkaHandler struct {
	templateRepo     *repositories.TemplateRepository
	renderJobRepo    *repositories.RenderJobRepository
	eventMappingRepo *repositories.EventMappingRepository
	kafkaLogRepo     *repositories.KafkaLogRepository
	renderer         *RendererService
	validator        *ValidatorService
	imageGen         *ImageGeneratorService
	logger           *logrus.Logger
}

func NewKafkaHandler(
	templateRepo *repositories.TemplateRepository,
	renderJobRepo *repositories.RenderJobRepository,
	eventMappingRepo *repositories.EventMappingRepository,
	kafkaLogRepo *repositories.KafkaLogRepository,
	renderer *RendererService,
	validator *ValidatorService,
	imageGen *ImageGeneratorService,
	logger *logrus.Logger,
) *KafkaHandler {
	return &KafkaHandler{
		templateRepo:     templateRepo,
		renderJobRepo:    renderJobRepo,
		eventMappingRepo: eventMappingRepo,
		kafkaLogRepo:     kafkaLogRepo,
		renderer:         renderer,
		validator:        validator,
		imageGen:         imageGen,
		logger:           logger,
	}
}

func (h *KafkaHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *KafkaHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *KafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.ProcessRawMessage(session.Context(), msg, "")
		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *KafkaHandler) ProcessRawMessage(ctx context.Context, msg *sarama.ConsumerMessage, connectionID string) {
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		h.logger.WithError(err).Error("Failed to unmarshal Kafka message")
		return
	}

	eventType := ""
	if et, ok := payload["event_type"].(string); ok {
		eventType = et
	}
	eventID := ""
	if eid, ok := payload["event_id"].(string); ok {
		eventID = eid
	}

	if eventType == "" && connectionID == "" {
		h.processRenderPayload(ctx, msg, payload, connectionID, eventID, eventType)
		return
	}

	if h.eventMappingRepo != nil && eventType != "" {
		h.processWithMapping(ctx, msg, payload, connectionID, eventType, eventID)
		return
	}

	h.processRenderPayload(ctx, msg, payload, connectionID, eventID, eventType)
}

func (h *KafkaHandler) processWithMapping(ctx context.Context, msg *sarama.ConsumerMessage, payload map[string]interface{}, connectionID string, eventType string, eventID string) {
	mapping, err := h.eventMappingRepo.GetByEventTypeAndConnection(ctx, eventType, connectionID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"event_type":    eventType,
			"connection_id": connectionID,
		}).Debug("No event mapping found, skipping")
		return
	}

	if !mapping.Active {
		return
	}

	logEntry := &models.KafkaLog{
		ConnectionID: connectionID,
		Topic:        msg.Topic,
		EventType:    eventType,
		EventID:      eventID,
		Status:       "processing",
		Payload:      payload,
		Partition:    msg.Partition,
		Offset:       msg.Offset,
	}

	data := extractData(payload)

	if len(mapping.FilterConditions) > 0 {
		passed, reason := EvaluateFilters(mapping.FilterConditions, data)
		if !passed {
			logEntry.Status = "filtered"
			logEntry.Message = reason
			h.saveLog(ctx, logEntry)
			return
		}
	}

	templateSlug := mapping.TemplateSlug
	if templateSlug == "" {
		logEntry.Status = "failed"
		logEntry.Error = "event mapping has no template_slug"
		h.saveLog(ctx, logEntry)
		return
	}

	h.renderAndLog(ctx, logEntry, templateSlug, data)
}

func (h *KafkaHandler) processRenderPayload(ctx context.Context, msg *sarama.ConsumerMessage, payload map[string]interface{}, connectionID string, eventID string, eventType string) {
	var rp models.RenderTriggerPayload
	payloadBytes, _ := json.Marshal(payload)
	_ = json.Unmarshal(payloadBytes, &rp)

	if rp.TemplateSlug == "" {
		return
	}

	logEntry := &models.KafkaLog{
		ConnectionID: connectionID,
		Topic:        msg.Topic,
		EventType:    eventType,
		EventID:      eventID,
		Status:       "processing",
		Payload:      payload,
		Partition:    msg.Partition,
		Offset:       msg.Offset,
	}

	data := rp.Data
	if data == nil {
		data = make(map[string]interface{})
		for k, v := range payload {
			if k != "template_slug" && k != "output_format" && k != "width" && k != "height" && k != "scale" && k != "callback_url" && k != "event_type" && k != "event_id" {
				data[k] = v
			}
		}
	}

	h.renderAndLog(ctx, logEntry, rp.TemplateSlug, data)
}

func (h *KafkaHandler) renderAndLog(ctx context.Context, logEntry *models.KafkaLog, templateSlug string, data map[string]interface{}) {
	tmpl, err := h.templateRepo.GetBySlug(ctx, templateSlug)
	if err != nil {
		logEntry.Status = "failed"
		logEntry.Error = fmt.Sprintf("template not found: %s", templateSlug)
		h.saveLog(ctx, logEntry)
		return
	}

	if !tmpl.IsActive {
		logEntry.Status = "failed"
		logEntry.Error = "template is not active"
		h.saveLog(ctx, logEntry)
		return
	}

	if tmpl.DataSchema != "" {
		valid, _, err := h.validator.ValidateData(tmpl.DataSchema, data)
		if err != nil || !valid {
			logEntry.Status = "failed"
			if err != nil {
				logEntry.Error = err.Error()
			} else {
				logEntry.Error = "data validation failed"
			}
			h.saveLog(ctx, logEntry)
			return
		}
	}

	mappedData := data
	if len(tmpl.FieldMapping) > 0 {
		mappedData = ApplyFieldMapping(data, tmpl.FieldMapping)
	}

	outputFormat := string(tmpl.OutputFormat)
	if outputFormat == "" {
		outputFormat = string(models.RenderFormatPNG)
	}

	job := &models.RenderJob{
		TemplateID:   tmpl.ID,
		TemplateSlug: tmpl.Slug,
		Data:         mappedData,
		OutputFormat: models.RenderOutputFormat(outputFormat),
		Width:        int(tmpl.Width),
		Height:       int(tmpl.Height),
		Status:       models.RenderStatusPending,
		Source:       "kafka",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_ = h.renderJobRepo.Create(ctx, job)

	job.Status = models.RenderStatusProcessing
	job.UpdatedAt = time.Now()
	_ = h.renderJobRepo.Update(ctx, job)

	var renderedHTML string
	if tmpl.BackgroundImage != "" {
		renderedHTML, err = h.renderer.RenderPositioned(tmpl, mappedData)
	} else {
		renderedHTML, err = h.renderer.RenderHTML(tmpl.HTML, tmpl.CSS, tmpl.Variables, mappedData)
	}
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		_ = h.renderJobRepo.Update(ctx, job)
		logEntry.Status = "failed"
		logEntry.Error = err.Error()
		h.saveLog(ctx, logEntry)
		return
	}

	opts := &models.RenderOptions{
		Width:   int(tmpl.Width),
		Height:  int(tmpl.Height),
		DPI:     tmpl.DPI,
		Format:  outputFormat,
		Quality: tmpl.Quality,
		Scale:   tmpl.Scale,
	}

	imgBytes, err := h.imageGen.GenerateFromHTML(ctx, renderedHTML, opts)
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		_ = h.renderJobRepo.Update(ctx, job)
		logEntry.Status = "failed"
		logEntry.Error = err.Error()
		h.saveLog(ctx, logEntry)
		return
	}

	now := time.Now()
	job.Status = models.RenderStatusCompleted
	job.OutputSize = len(imgBytes)
	job.Duration = now.Unix() - job.CreatedAt.Unix()
	job.UpdatedAt = now
	job.CompletedAt = &now
	_ = h.renderJobRepo.Update(ctx, job)

	logEntry.Status = "processed"
	logEntry.RenderJobID = job.ID.Hex()
	logEntry.Message = fmt.Sprintf("rendered %s (%d bytes)", templateSlug, len(imgBytes))
	h.saveLog(ctx, logEntry)

	h.logger.WithFields(logrus.Fields{
		"job_id":     job.ID,
		"template":   tmpl.Slug,
		"event_type": logEntry.EventType,
		"size":       len(imgBytes),
	}).Info("Render job completed via Kafka event mapping")
}

func (h *KafkaHandler) saveLog(ctx context.Context, logEntry *models.KafkaLog) {
	if h.kafkaLogRepo != nil {
		_ = h.kafkaLogRepo.Create(ctx, logEntry)
	}
}

func extractData(payload map[string]interface{}) map[string]interface{} {
	if data, ok := payload["data"].(map[string]interface{}); ok {
		return data
	}
	result := make(map[string]interface{})
	for k, v := range payload {
		if k != "event_type" && k != "event_id" {
			result[k] = v
		}
	}
	return result
}
