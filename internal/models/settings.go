package models

import "time"

type Settings struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Key       string    `bson:"key" json:"key"`
	Value     string    `bson:"value" json:"value"`
	Category  string    `bson:"category,omitempty" json:"category,omitempty"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
