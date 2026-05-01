package models

type RenderRequest struct {
	TemplateSlug string                 `json:"template_slug" validate:"required"`
	Data         map[string]interface{} `json:"data"`
	OutputFormat string                 `json:"output_format,omitempty"`
	Mode         string                 `json:"mode,omitempty"`
	Width        int                    `json:"width,omitempty"`
	Height       int                    `json:"height,omitempty"`
	Quality      int                    `json:"quality,omitempty"`
	Async        bool                   `json:"async,omitempty"`
}

type BatchRenderRequest struct {
	TemplateSlug string                   `json:"template_slug" validate:"required"`
	Items        []map[string]interface{} `json:"items" validate:"required,min=1"`
	Format       string                   `json:"format,omitempty"`
	Width        int                      `json:"width,omitempty"`
	Height       int                      `json:"height,omitempty"`
	Quality      int                      `json:"quality,omitempty"`
}

type RenderResponse struct {
	Success      bool   `json:"success"`
	JobID        string `json:"job_id,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	ImageID      string `json:"image_id,omitempty"`
	TemplateUsed string `json:"template_used,omitempty"`
	Error        string `json:"error,omitempty"`
}

type BatchRenderResponse struct {
	Success bool             `json:"success"`
	JobIDs  []string         `json:"job_ids,omitempty"`
	Results []RenderResponse `json:"results,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type CreateTemplateRequest struct {
	Name           string             `json:"name" validate:"required"`
	Slug           string             `json:"slug" validate:"required"`
	Description    string             `json:"description,omitempty"`
	HTML           string             `json:"html,omitempty"`
	CSS            string             `json:"css,omitempty"`
	DataSchema     string             `json:"data_schema,omitempty"`
	Variables      []TemplateVariable `json:"variables,omitempty"`
	Width          float64            `json:"width" validate:"required"`
	Height         float64            `json:"height" validate:"required"`
	DimensionUnit  DimensionUnit      `json:"dimension_unit,omitempty"`
	DPI            int                `json:"dpi,omitempty"`
	OutputFormat   string             `json:"output_format,omitempty"`
	Quality        int                `json:"quality,omitempty"`
	Tags           []Tag              `json:"tags,omitempty"`
	IsActive       *bool              `json:"is_active,omitempty"`
	BackgroundImage *string           `json:"background_image,omitempty"`
	FieldMapping   []FieldMapping     `json:"field_mapping,omitempty"`
}

type UpdateTemplateRequest struct {
	Name           *string            `json:"name,omitempty"`
	Description    *string            `json:"description,omitempty"`
	HTML           *string            `json:"html,omitempty"`
	CSS            *string            `json:"css,omitempty"`
	DataSchema     *string            `json:"data_schema,omitempty"`
	Variables      []TemplateVariable `json:"variables,omitempty"`
	BackgroundImage *string           `json:"background_image,omitempty"`
	FieldMapping   []FieldMapping     `json:"field_mapping,omitempty"`
	Width          *float64           `json:"width,omitempty"`
	Height         *float64           `json:"height,omitempty"`
	DimensionUnit  *DimensionUnit     `json:"dimension_unit,omitempty"`
	DPI            *int               `json:"dpi,omitempty"`
	OutputFormat   *string            `json:"output_format,omitempty"`
	Quality        *int               `json:"quality,omitempty"`
	Tags           []Tag              `json:"tags,omitempty"`
	IsActive       *bool              `json:"is_active,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
