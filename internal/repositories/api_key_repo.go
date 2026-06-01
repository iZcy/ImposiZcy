package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const apiKeysCollection = "api_keys"

type APIKeyRepository struct {
	db *mongo.Database
}

func NewAPIKeyRepository(db *mongo.Database) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) collection() *mongo.Collection {
	return r.db.Collection(apiKeysCollection)
}

func (r *APIKeyRepository) Create(ctx context.Context, name, createdBy string) (*models.APIKey, string, error) {
	rawKey := generateAPIKey()
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])
	prefix := rawKey[:8]

	apiKey := &models.APIKey{
		Name:      name,
		KeyHash:   keyHash,
		Prefix:    prefix,
		Active:    true,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	_, err := r.collection().InsertOne(ctx, apiKey)
	if err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (r *APIKeyRepository) Validate(ctx context.Context, rawKey string) (*models.APIKey, error) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var apiKey models.APIKey
	err := r.collection().FindOne(ctx, bson.M{"key_hash": keyHash, "active": true}).Decode(&apiKey)
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *APIKeyRepository) List(ctx context.Context) ([]*models.APIKey, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []*models.APIKey
	if err := cursor.All(ctx, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"active": false}})
	return err
}

func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
