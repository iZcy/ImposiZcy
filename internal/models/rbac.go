package models

import "time"

type InternalRole struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	Username    string    `bson:"username" json:"username"`
	Role        string    `bson:"role" json:"role"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

type ExternalRole struct {
	ID          string    `bson:"_id,omitempty" json:"id,omitempty"`
	ServiceName string    `bson:"service_name" json:"service_name"`
	Role        string    `bson:"role" json:"role"`
	Permissions []string  `bson:"permissions" json:"permissions"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}
