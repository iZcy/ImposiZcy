package models

import "time"

type FilterCondition struct {
	Field    string `json:"field" bson:"field"`
	Operator string `json:"operator" bson:"operator"`
	Value    string `json:"value" bson:"value"`
}

type EventMapping struct {
	ID               string             `json:"id" bson:"_id,omitempty"`
	EventType        string             `json:"event_type" bson:"event_type"`
	TemplateID       string             `json:"template_id" bson:"template_id"`
	TemplateSlug     string             `json:"template_slug" bson:"template_slug"`
	ConnectionID     string             `json:"connection_id" bson:"connection_id"`
	Description      string             `json:"description" bson:"description"`
	Active           bool               `json:"active" bson:"active"`
	FilterConditions []FilterCondition  `json:"filter_conditions,omitempty" bson:"filter_conditions,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

type EventMappingRequest struct {
	EventType        string             `json:"event_type" binding:"required"`
	TemplateSlug     string             `json:"template_slug" binding:"required"`
	ConnectionID     string             `json:"connection_id"`
	Description      string             `json:"description"`
	Active           *bool              `json:"active"`
	FilterConditions []FilterCondition  `json:"filter_conditions,omitempty"`
}

type KafkaLog struct {
	ID           string                 `bson:"_id,omitempty" json:"id"`
	ConnectionID string                 `bson:"connection_id" json:"connection_id"`
	Topic        string                 `bson:"topic" json:"topic"`
	EventType    string                 `bson:"event_type" json:"event_type"`
	EventID      string                 `bson:"event_id" json:"event_id"`
	Status       string                 `bson:"status" json:"status"`
	Message      string                 `bson:"message,omitempty" json:"message,omitempty"`
	Error        string                 `bson:"error,omitempty" json:"error,omitempty"`
	Payload      map[string]interface{} `bson:"payload" json:"payload"`
	RenderJobID  string                 `bson:"render_job_id,omitempty" json:"render_job_id,omitempty"`
	Partition    int32                  `bson:"partition" json:"partition"`
	Offset       int64                  `bson:"offset" json:"offset"`
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
}

type KafkaLogFilter struct {
	Status       string `form:"status"`
	Topic        string `form:"topic"`
	ConnectionID string `form:"connection_id"`
	EventType    string `form:"event_type"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
}

type KafkaLogListResponse struct {
	Success bool        `json:"success"`
	Data    []*KafkaLog `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
}
