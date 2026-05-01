package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const kafkaLogsCollection = "kafka_logs"

type KafkaLogRepository struct {
	db *mongo.Database
}

func NewKafkaLogRepository(db *mongo.Database) *KafkaLogRepository {
	return &KafkaLogRepository{db: db}
}

func (r *KafkaLogRepository) collection() *mongo.Collection {
	return r.db.Collection(kafkaLogsCollection)
}

func (r *KafkaLogRepository) Create(ctx context.Context, log *models.KafkaLog) error {
	log.ID = primitive.NewObjectID().Hex()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	_, err := r.collection().InsertOne(ctx, log)
	return err
}

func (r *KafkaLogRepository) List(ctx context.Context, filter models.KafkaLogFilter) ([]*models.KafkaLog, int64, error) {
	query := bson.M{}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.Topic != "" {
		query["topic"] = filter.Topic
	}
	if filter.ConnectionID != "" {
		query["connection_id"] = filter.ConnectionID
	}
	if filter.EventType != "" {
		query["event_type"] = filter.EventType
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total, _ := r.collection().CountDocuments(ctx, query)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.collection().Find(ctx, query, opts)
	if err != nil {
		return []*models.KafkaLog{}, 0, nil
	}
	defer cursor.Close(ctx)

	var logs []*models.KafkaLog
	if err := cursor.All(ctx, &logs); err != nil {
		return []*models.KafkaLog{}, 0, nil
	}
	if logs == nil {
		logs = []*models.KafkaLog{}
	}
	return logs, total, nil
}

func (r *KafkaLogRepository) DeleteAll(ctx context.Context) error {
	_, err := r.collection().DeleteMany(ctx, bson.M{})
	return err
}

func (r *KafkaLogRepository) DeleteOlderThan(ctx context.Context, age time.Duration) error {
	cutoff := time.Now().Add(-age)
	_, err := r.collection().DeleteMany(ctx, bson.M{"created_at": bson.M{"$lt": cutoff}})
	return err
}
