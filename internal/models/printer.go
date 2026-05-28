package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PrinterStatus string

const (
	PrinterStatusIdle     PrinterStatus = "idle"
	PrinterStatusPrinting PrinterStatus = "printing"
	PrinterStatusError    PrinterStatus = "error"
	PrinterStatusOffline  PrinterStatus = "offline"
)

type Printer struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name       string             `bson:"name" json:"name" validate:"required"`
	Location   string             `bson:"location" json:"location"`
	CupsName   string             `bson:"cups_name" json:"cups_name" validate:"required"`
	Status     PrinterStatus      `bson:"status" json:"status"`
	PaperSizes []string           `bson:"paper_sizes" json:"paper_sizes"`
	ColorModes []string           `bson:"color_modes" json:"color_modes"`
	IsActive   bool               `bson:"is_active" json:"is_active"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreatePrinterRequest struct {
	Name       string   `json:"name" validate:"required"`
	Location   string   `json:"location,omitempty"`
	CupsName   string   `json:"cups_name" validate:"required"`
	PaperSizes []string `json:"paper_sizes,omitempty"`
	ColorModes []string `json:"color_modes,omitempty"`
	IsActive   bool     `json:"is_active,omitempty"`
}

type UpdatePrinterRequest struct {
	Name       *string  `json:"name,omitempty"`
	Location   *string  `json:"location,omitempty"`
	CupsName   *string  `json:"cups_name,omitempty"`
	PaperSizes []string `json:"paper_sizes,omitempty"`
	ColorModes []string `json:"color_modes,omitempty"`
	IsActive   *bool    `json:"is_active,omitempty"`
	Status     *string  `json:"status,omitempty"`
}
