package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tag struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

type DimensionUnit string

const (
	DimensionUnitPX   DimensionUnit = "px"
	DimensionUnitMM   DimensionUnit = "mm"
	DimensionUnitCM   DimensionUnit = "cm"
	DimensionUnitInch DimensionUnit = "inch"
)

type TemplateStatus string

const (
	TemplateStatusDraft     TemplateStatus = "draft"
	TemplateStatusPublished TemplateStatus = "published"
	TemplateStatusArchived  TemplateStatus = "archived"
)

type OutputFormatType string

const (
	OutputFormatPNG  OutputFormatType = "png"
	OutputFormatJPEG OutputFormatType = "jpeg"
	OutputFormatWEBP OutputFormatType = "webp"
)

type TemplateVariableType string

const (
	VariableTypeText    TemplateVariableType = "text"
	VariableTypeBarcode TemplateVariableType = "barcode"
	VariableTypeImage   TemplateVariableType = "image"
)

type BarcodeFormat string

const (
	BarcodeFormatCode128 BarcodeFormat = "code128"
	BarcodeFormatQR      BarcodeFormat = "qr"
	BarcodeFormatEAN13   BarcodeFormat = "ean13"
	BarcodeFormatEAN8    BarcodeFormat = "ean8"
	BarcodeFormatUPCA    BarcodeFormat = "upca"
	BarcodeFormatCode39  BarcodeFormat = "code39"
)

// FieldPosition defines the placement of a variable on the template (DocuSign-style)
type FieldPosition struct {
	X         float64 `bson:"x,omitempty" json:"x,omitempty"`                   // X position in px (from left)
	Y         float64 `bson:"y,omitempty" json:"y,omitempty"`                   // Y position in px (from top)
	Width     float64 `bson:"width,omitempty" json:"width,omitempty"`           // Width in px
	Height    float64 `bson:"height,omitempty" json:"height,omitempty"`         // Height in px
	FontSize  float64 `bson:"font_size,omitempty" json:"font_size,omitempty"`   // Font size in px
	Alignment string  `bson:"alignment,omitempty" json:"alignment,omitempty"`   // left, center, right
	Color     string  `bson:"color,omitempty" json:"color,omitempty"`           // Hex color (e.g. #000000)
	Bold      bool    `bson:"bold,omitempty" json:"bold,omitempty"`             // Bold text
}

// FieldMapping maps an external source field to a template variable name.
// This allows incoming data (e.g. from Bracelet Service) with keys like "ticket.barcode"
// to be mapped to template variables like "barcode".
type FieldMapping struct {
	SourceField   string `bson:"source_field" json:"source_field" validate:"required"`     // Key in incoming data
	TargetVariable string `bson:"target_variable" json:"target_variable" validate:"required"` // Template variable name
	DefaultValue  string `bson:"default_value,omitempty" json:"default_value,omitempty"`    // Fallback if source is empty
	Transform     string `bson:"transform,omitempty" json:"transform,omitempty"`            // Optional: "uppercase", "lowercase", "trim"
}

type TemplateVariable struct {
	Name          string               `bson:"name" json:"name" validate:"required"`
	Type          TemplateVariableType `bson:"type" json:"type" validate:"required"`
	BarcodeFormat BarcodeFormat        `bson:"barcode_format,omitempty" json:"barcode_format,omitempty"`
	Required      bool                 `bson:"required" json:"required"`
	DefaultValue  string               `bson:"default_value,omitempty" json:"default_value,omitempty"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty"`
	// Position-based placement (DocuSign-style positioning on background image)
	Position    *FieldPosition `bson:"position,omitempty" json:"position,omitempty"`
	FontSize    float64        `bson:"font_size,omitempty" json:"font_size,omitempty"`       // Font size in px (default 14)
	FontColor   string         `bson:"font_color,omitempty" json:"font_color,omitempty"`     // Hex color (default #000000)
	FontWeight  string         `bson:"font_weight,omitempty" json:"font_weight,omitempty"`   // "normal" or "bold"
	TextAlign   string         `bson:"text_align,omitempty" json:"text_align,omitempty"`     // "left", "center", "right"
	SourceField string         `bson:"source_field,omitempty" json:"source_field,omitempty"` // Maps external field name to this variable
}

type PrintTemplate struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name" json:"name" validate:"required"`
	Slug            string             `bson:"slug" json:"slug" validate:"required"`
	Description     string             `bson:"description,omitempty" json:"description,omitempty"`
	HTML            string             `bson:"html" json:"html,omitempty"`
	CSS             string             `bson:"css,omitempty" json:"css,omitempty"`
	DataSchema      string             `bson:"data_schema,omitempty" json:"data_schema,omitempty"`
	Variables       []TemplateVariable `bson:"variables,omitempty" json:"variables,omitempty"`
	FieldMapping    []FieldMapping     `bson:"field_mapping,omitempty" json:"field_mapping,omitempty"`
	BackgroundImage string             `bson:"background_image,omitempty" json:"background_image,omitempty"`
	Width           float64            `bson:"width" json:"width" validate:"required"`
	Height          float64            `bson:"height" json:"height" validate:"required"`
	DimensionUnit   DimensionUnit      `bson:"dimension_unit,omitempty" json:"dimension_unit,omitempty"`
	DPI             int                `bson:"dpi,omitempty" json:"dpi,omitempty"`
	OutputFormat    OutputFormatType   `bson:"output_format,omitempty" json:"output_format,omitempty"`
	Quality         int                `bson:"quality,omitempty" json:"quality,omitempty"`
	Scale           float64            `bson:"scale,omitempty" json:"scale,omitempty"`
	Tags            []Tag              `bson:"tags,omitempty" json:"tags,omitempty"`
	IsActive        bool               `bson:"is_active" json:"is_active"`
	Status          TemplateStatus     `bson:"status,omitempty" json:"status,omitempty"`
	Orientation     string             `bson:"orientation,omitempty" json:"orientation,omitempty"`
	Background      string             `bson:"background,omitempty" json:"background,omitempty"`
	PreviewImage    string             `bson:"preview_image,omitempty" json:"preview_image,omitempty"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
