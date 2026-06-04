package models

import "time"

type APIKey struct {
	ID        string     `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string     `bson:"name" json:"name"`
	TenantID  string     `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	KeyHash   string     `bson:"key_hash" json:"-"`
	RawKey    string     `bson:"raw_key,omitempty" json:"-"` // Stored for tenant key retrieval (less secure but necessary)
	Prefix    string     `bson:"prefix" json:"prefix"`
	Active    bool       `bson:"active" json:"active"`
	CreatedBy string     `bson:"created_by,omitempty" json:"created_by,omitempty"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	ExpiresAt *time.Time `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}
