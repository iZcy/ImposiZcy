package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FontScope string

const (
	FontScopeGlobal   FontScope = "global"
	FontScopeTemplate FontScope = "template"
)

type FontFormat string

const (
	FontFormatTTF FontFormat = "ttf"
	FontFormatOTF FontFormat = "otf"
)

type Font struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Family     string             `bson:"family" json:"family" validate:"required"`
	FileName   string             `bson:"file_name" json:"file_name"`
	Path       string             `bson:"path" json:"path"`
	Format     FontFormat         `bson:"format" json:"format"`
	Scope      FontScope          `bson:"scope" json:"scope"`
	TemplateID string             `bson:"template_id,omitempty" json:"template_id,omitempty"`
	Size       int64              `bson:"size" json:"size"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
}
