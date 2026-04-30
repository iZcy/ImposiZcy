package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const settingsCollection = "settings"

type SettingsRepository struct {
	db *mongo.Database
}

func NewSettingsRepository(db *mongo.Database) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) collection() *mongo.Collection {
	return r.db.Collection(settingsCollection)
}

func (r *SettingsRepository) Get(ctx context.Context, key string) (*models.Settings, error) {
	var setting models.Settings
	err := r.collection().FindOne(ctx, bson.M{"key": key}).Decode(&setting)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key, value string, category ...string) error {
	cat := ""
	if len(category) > 0 {
		cat = category[0]
	}
	_, err := r.collection().UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{"$set": bson.M{
			"key":        key,
			"value":      value,
			"category":   cat,
			"updated_at": time.Now(),
		}, "$setOnInsert": bson.M{
			"created_at": time.Now(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *SettingsRepository) GetAll(ctx context.Context) ([]*models.Settings, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settings []*models.Settings
	if err := cursor.All(ctx, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingsRepository) GetDashboardIPWhitelist(ctx context.Context) []string {
	setting, err := r.Get(ctx, "dashboard_ip_whitelist")
	if err != nil {
		return nil
	}
	if setting == nil || setting.Value == "" {
		return nil
	}
	return splitAndTrim(setting.Value)
}

func (r *SettingsRepository) GetAPIIPWhitelist(ctx context.Context) []string {
	setting, err := r.Get(ctx, "api_ip_whitelist")
	if err != nil {
		return nil
	}
	if setting == nil || setting.Value == "" {
		return nil
	}
	return splitAndTrim(setting.Value)
}

func splitAndTrim(s string) []string {
	var result []string
	for _, v := range splitString(s, ",") {
		if trimmed := trimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (r *SettingsRepository) GetDashboardIPWhitelistNoCtx() []string {
	setting, err := r.Get(context.Background(), "dashboard_ip_whitelist")
	if err != nil {
		return nil
	}
	if setting == nil || setting.Value == "" {
		return nil
	}
	return splitAndTrim(setting.Value)
}

func (r *SettingsRepository) GetAPIIPWhitelistNoCtx() []string {
	setting, err := r.Get(context.Background(), "api_ip_whitelist")
	if err != nil {
		return nil
	}
	if setting == nil || setting.Value == "" {
		return nil
	}
	return splitAndTrim(setting.Value)
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
