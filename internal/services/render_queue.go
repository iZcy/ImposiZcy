package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type RenderQueue struct {
	templateRepo *repositories.TemplateRepository
	imageRepo    *repositories.ImageOutputRepository
	renderer     *RendererService
	imageGen     *ImageGeneratorService
	validator    *ValidatorService
	logger       *logrus.Logger
	outputDir    string
}

func NewRenderQueue(
	templateRepo *repositories.TemplateRepository,
	imageRepo *repositories.ImageOutputRepository,
	renderer *RendererService,
	imageGen *ImageGeneratorService,
	validator *ValidatorService,
	logger *logrus.Logger,
	outputDir string,
) *RenderQueue {
	os.MkdirAll(outputDir, 0755)
	return &RenderQueue{
		templateRepo: templateRepo,
		imageRepo:    imageRepo,
		renderer:     renderer,
		imageGen:     imageGen,
		validator:    validator,
		logger:       logger,
		outputDir:    outputDir,
	}
}

func (q *RenderQueue) ProcessJob(ctx context.Context, job *models.RenderJob) (*models.ImageOutput, error) {
	q.logger.WithFields(logrus.Fields{
		"job_id":        job.ID,
		"template_slug": job.TemplateSlug,
	}).Info("Processing render job")

	tmpl, err := q.templateRepo.GetBySlug(ctx, job.TemplateSlug)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	if tmpl.DataSchema != "" {
		valid, errs, valErr := q.validator.ValidateData(tmpl.DataSchema, job.Data)
		if valErr != nil {
			return nil, fmt.Errorf("validation error: %w", valErr)
		}
		if !valid {
			return nil, fmt.Errorf("data validation failed: %v", errs)
		}
	}

	html, err := q.renderer.RenderHTML(tmpl.HTML, job.Data)
	if err != nil {
		return nil, fmt.Errorf("render failed: %w", err)
	}

	width := tmpl.Width
	height := tmpl.Height
	format := string(job.OutputFormat)
	if format == "" {
		format = "png"
	}

	opts := &models.RenderOptions{
		Width:  int(width),
		Height: int(height),
		Format: format,
		Scale:  2.0,
	}

	imageBytes, err := q.imageGen.GenerateFromHTML(ctx, html, opts)
	if err != nil {
		return nil, fmt.Errorf("image generation failed: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.%s", job.TemplateSlug, uuid.New().String()[:8], format)
	filePath := filepath.Join(q.outputDir, filename)
	if err := os.WriteFile(filePath, imageBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	now := time.Now()
	output := &models.ImageOutput{
		TemplateID:   tmpl.ID.Hex(),
		TemplateSlug: tmpl.Slug,
		Format:       models.OutputFormat(format),
		Width:        int(width),
		Height:       int(height),
		Data:         job.Data,
		FilePath:     filePath,
		FileSize:     len(imageBytes),
		Status:       models.RenderStatusCompleted,
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := q.imageRepo.Create(ctx, output); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save output metadata: %w", err)
	}

	q.logger.WithFields(logrus.Fields{
		"job_id":    job.ID,
		"output_id": output.ID,
		"filename":  filename,
	}).Info("Render job completed")

	return output, nil
}
