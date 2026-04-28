package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"

	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
)

type KafkaHandler struct {
	templateRepo  *repositories.TemplateRepository
	renderJobRepo *repositories.RenderJobRepository
	renderer      *RendererService
	validator     *ValidatorService
	imageGen      *ImageGeneratorService
	logger        *logrus.Logger
}

func NewKafkaHandler(
	templateRepo *repositories.TemplateRepository,
	renderJobRepo *repositories.RenderJobRepository,
	renderer *RendererService,
	validator *ValidatorService,
	imageGen *ImageGeneratorService,
	logger *logrus.Logger,
) *KafkaHandler {
	return &KafkaHandler{
		templateRepo:  templateRepo,
		renderJobRepo: renderJobRepo,
		renderer:      renderer,
		validator:     validator,
		imageGen:      imageGen,
		logger:        logger,
	}
}

func (h *KafkaHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *KafkaHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *KafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.logger.WithFields(logrus.Fields{
			"topic":     msg.Topic,
			"partition": msg.Partition,
			"offset":    msg.Offset,
		}).Info("Processing Kafka message")

		if err := h.processMessage(session.Context(), msg); err != nil {
			h.logger.WithError(err).Error("Failed to process Kafka message")
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

func (h *KafkaHandler) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	var payload models.RenderTriggerPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	tmpl, err := h.templateRepo.GetBySlug(ctx, payload.TemplateSlug)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	if tmpl.DataSchema != "" {
		valid, validationErrors, err := h.validator.ValidateData(tmpl.DataSchema, payload.Data)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
		if !valid {
			return fmt.Errorf("data validation failed: %v", validationErrors)
		}
	}

	outputFormat := payload.OutputFormat
	if outputFormat == "" {
		outputFormat = models.OutputFormat(tmpl.OutputFormat)
	}
	if outputFormat == "" {
		outputFormat = models.OutputFormatPNG
	}

	job := &models.RenderJob{
		TemplateID:   tmpl.ID,
		TemplateSlug: tmpl.Slug,
		Data:         payload.Data,
		OutputFormat: outputFormat,
		Status:       models.RenderStatusPending,
		Source:       "kafka",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.renderJobRepo.Create(ctx, job); err != nil {
		return fmt.Errorf("failed to create render job: %w", err)
	}

	job.Status = models.RenderStatusProcessing
	job.UpdatedAt = time.Now()
	_ = h.renderJobRepo.Update(ctx, job)

	html, err := h.renderer.RenderHTML(tmpl.HTML, payload.Data)
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		_ = h.renderJobRepo.Update(ctx, job)
		return err
	}

	opts := &models.RenderOptions{
		Width:   int(tmpl.Width),
		Height:  int(tmpl.Height),
		DPI:     tmpl.DPI,
		Format:  string(outputFormat),
		Quality: tmpl.Quality,
		Scale:   tmpl.Scale,
	}

	imgBytes, err := h.imageGen.GenerateFromHTML(ctx, html, opts)
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		job.UpdatedAt = time.Now()
		_ = h.renderJobRepo.Update(ctx, job)
		return err
	}

	now := time.Now()
	job.Status = models.RenderStatusCompleted
	job.OutputSize = len(imgBytes)
	job.Duration = now.Unix() - job.CreatedAt.Unix()
	job.UpdatedAt = now
	completedAt := now
	job.CompletedAt = &completedAt
	_ = h.renderJobRepo.Update(ctx, job)

	h.logger.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"template": tmpl.Slug,
		"size":     len(imgBytes),
	}).Info("Render job completed via Kafka")

	return nil
}
