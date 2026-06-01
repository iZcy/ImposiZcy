package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PrintJobStatus string

const (
	PrintJobStatusPending   PrintJobStatus = "pending"
	PrintJobStatusPrinting  PrintJobStatus = "printing"
	PrintJobStatusDone      PrintJobStatus = "done"
	PrintJobStatusFailed    PrintJobStatus = "failed"
	PrintJobStatusCancelled PrintJobStatus = "cancelled"
)

type PrintJob struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TenantID    string             `bson:"tenant_id" json:"tenant_id"`
	OrderID     string             `bson:"order_id" json:"order_id" validate:"required"`
	PrinterID   string             `bson:"printer_id" json:"printer_id" validate:"required"`
	FileURL     string             `bson:"file_url" json:"file_url"`
	FilePath    string             `bson:"file_path,omitempty" json:"file_path,omitempty"`
	Status      PrintJobStatus     `bson:"status" json:"status"`
	Copies      int                `bson:"copies" json:"copies"`
	PaperSize   string             `bson:"paper_size" json:"paper_size"`
	ColorMode   string             `bson:"color_mode" json:"color_mode"`
	Pages       int                `bson:"pages" json:"pages"`
	Sides       string             `bson:"sides" json:"sides"`
	CUPSJobID   string             `bson:"cups_job_id,omitempty" json:"cups_job_id,omitempty"`
	CallbackURL string             `bson:"callback_url,omitempty" json:"callback_url,omitempty"`
	ErrorMsg    string             `bson:"error_msg,omitempty" json:"error_msg,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type CreatePrintJobRequest struct {
	TenantID    string `json:"tenant_id,omitempty"`
	OrderID     string `json:"order_id" validate:"required"`
	PrinterID   string `json:"printer_id" validate:"required"`
	FileURL     string `json:"file_url,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	Copies      int    `json:"copies,omitempty"`
	PaperSize   string `json:"paper_size,omitempty"`
	ColorMode   string `json:"color_mode,omitempty"`
	Pages       int    `json:"pages,omitempty"`
	Sides       string `json:"sides,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

type UpdatePrintJobStatusRequest struct {
	Status   string `json:"status" validate:"required"`
	ErrorMsg string `json:"error_msg,omitempty"`
}
