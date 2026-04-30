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

type TemplateVariable struct {
	Name          string               `bson:"name" json:"name" validate:"required"`
	Type          TemplateVariableType `bson:"type" json:"type" validate:"required"`
	BarcodeFormat BarcodeFormat        `bson:"barcode_format,omitempty" json:"barcode_format,omitempty"`
	Required      bool                 `bson:"required" json:"required"`
	DefaultValue  string               `bson:"default_value,omitempty" json:"default_value,omitempty"`
	Description   string               `bson:"description,omitempty" json:"description,omitempty"`
}

type PrintTemplate struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name" validate:"required"`
	Slug          string             `bson:"slug" json:"slug" validate:"required"`
	Description   string             `bson:"description,omitempty" json:"description,omitempty"`
	HTML          string             `bson:"html" json:"html,omitempty"`
	CSS           string             `bson:"css,omitempty" json:"css,omitempty"`
	DataSchema    string             `bson:"data_schema,omitempty" json:"data_schema,omitempty"`
	Variables     []TemplateVariable `bson:"variables,omitempty" json:"variables,omitempty"`
	Width         float64            `bson:"width" json:"width" validate:"required"`
	Height        float64            `bson:"height" json:"height" validate:"required"`
	DimensionUnit DimensionUnit      `bson:"dimension_unit,omitempty" json:"dimension_unit,omitempty"`
	DPI           int                `bson:"dpi,omitempty" json:"dpi,omitempty"`
	OutputFormat  OutputFormatType   `bson:"output_format,omitempty" json:"output_format,omitempty"`
	Quality       int                `bson:"quality,omitempty" json:"quality,omitempty"`
	Scale         float64            `bson:"scale,omitempty" json:"scale,omitempty"`
	Tags          []Tag              `bson:"tags,omitempty" json:"tags,omitempty"`
	IsActive      bool               `bson:"is_active" json:"is_active"`
	Status        TemplateStatus     `bson:"status,omitempty" json:"status,omitempty"`
	Orientation   string             `bson:"orientation,omitempty" json:"orientation,omitempty"`
	Background    string             `bson:"background,omitempty" json:"background,omitempty"`
	PreviewImage  string             `bson:"preview_image,omitempty" json:"preview_image,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}
