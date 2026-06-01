package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RenderJobStatus string

const (
	RenderStatusPending    RenderJobStatus = "pending"
	RenderStatusProcessing RenderJobStatus = "processing"
	RenderStatusCompleted  RenderJobStatus = "completed"
	RenderStatusFailed     RenderJobStatus = "failed"
)

type RenderOutputFormat string

const (
	RenderFormatPNG  RenderOutputFormat = "png"
	RenderFormatJPEG RenderOutputFormat = "jpeg"
	RenderFormatWEBP RenderOutputFormat = "webp"
)

type RenderJob struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	TemplateID    primitive.ObjectID     `bson:"template_id" json:"template_id"`
	TemplateSlug  string                 `bson:"template_slug" json:"template_slug"`
	Data          map[string]interface{} `bson:"data" json:"data"`
	Format        RenderOutputFormat     `bson:"format" json:"format"`
	OutputFormat  RenderOutputFormat     `bson:"output_format" json:"output_format"`
	Width         int                    `bson:"width" json:"width"`
	Height        int                    `bson:"height" json:"height"`
	Scale         float64                `bson:"scale" json:"scale"`
	Status        RenderJobStatus        `bson:"status" json:"status"`
	Source        string                 `bson:"source,omitempty" json:"source,omitempty"`
	OutputPath    string                 `bson:"output_path,omitempty" json:"output_path,omitempty"`
	OutputImageID string                 `bson:"output_image_id,omitempty" json:"output_image_id,omitempty"`
	OutputSize    int                    `bson:"output_size,omitempty" json:"output_size,omitempty"`
	Error         string                 `bson:"error,omitempty" json:"error,omitempty"`
	Duration      int64                  `bson:"duration,omitempty" json:"duration,omitempty"`
	ImageSizeBytes int64                 `bson:"image_size_bytes,omitempty" json:"image_size_bytes,omitempty"`
	CreatedAt     time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at" json:"updated_at"`
	CompletedAt   *time.Time             `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type ImageOutput struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	RenderJobID  string                 `bson:"render_job_id" json:"render_job_id"`
	TemplateID   string                 `bson:"template_id" json:"template_id"`
	TemplateName string                 `bson:"template_name,omitempty" json:"template_name,omitempty"`
	TemplateSlug string                 `bson:"template_slug" json:"template_slug"`
	Format       RenderOutputFormat     `bson:"format" json:"format"`
	Width        int                    `bson:"width" json:"width"`
	Height       int                    `bson:"height" json:"height"`
	DPI          int                    `bson:"dpi,omitempty" json:"dpi,omitempty"`
	FileSize     int                    `bson:"file_size" json:"file_size"`
	Data         map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
	Base64       string                 `bson:"-" json:"base64,omitempty"`
	FilePath     string                 `bson:"file_path,omitempty" json:"file_path,omitempty"`
	Filename     string                 `bson:"filename,omitempty" json:"filename,omitempty"`
	Status       RenderJobStatus        `bson:"status" json:"status"`
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
	CompletedAt  *time.Time             `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type RenderTriggerPayload struct {
	TemplateSlug string                 `json:"template_slug"`
	Data         map[string]interface{} `json:"data"`
	OutputFormat RenderOutputFormat     `json:"output_format,omitempty"`
	Width        int                    `json:"width,omitempty"`
	Height       int                    `json:"height,omitempty"`
	Scale        float64                `json:"scale,omitempty"`
	CallbackURL  string                 `json:"callback_url,omitempty"`
}

type RenderOptions struct {
	Width   int
	Height  int
	Format  string
	Quality int
	Scale   float64
	DPI     int
}
