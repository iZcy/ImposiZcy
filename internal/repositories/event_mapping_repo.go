package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const eventMappingsCollection = "event_mappings"

type EventMappingRepository struct {
	db *mongo.Database
}

func NewEventMappingRepository(db *mongo.Database) *EventMappingRepository {
	return &EventMappingRepository{db: db}
}

func (r *EventMappingRepository) collection() *mongo.Collection {
	return r.db.Collection(eventMappingsCollection)
}

func (r *EventMappingRepository) GetAll(ctx context.Context) ([]*models.EventMapping, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []*models.EventMapping
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *EventMappingRepository) GetByID(ctx context.Context, id string) (*models.EventMapping, error) {
	var mapping models.EventMapping
	err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *EventMappingRepository) GetByEventType(ctx context.Context, eventType string) (*models.EventMapping, error) {
	var mapping models.EventMapping
	err := r.collection().FindOne(ctx, bson.M{"event_type": eventType, "active": true}).Decode(&mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *EventMappingRepository) GetByEventTypeAndConnection(ctx context.Context, eventType string, connectionID string) (*models.EventMapping, error) {
	var filter bson.M
	if connectionID == "" {
		filter = bson.M{
			"event_type": eventType,
			"active":     true,
			"$or": []bson.M{
				{"connection_id": ""},
				{"connection_id": bson.M{"$exists": false}},
			},
		}
	} else {
		filter = bson.M{
			"event_type":    eventType,
			"connection_id": connectionID,
			"active":        true,
		}
	}

	var mapping models.EventMapping
	err := r.collection().FindOne(ctx, filter).Decode(&mapping)
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *EventMappingRepository) GetAllByEventTypeAndConnection(ctx context.Context, eventType string, connectionID string) ([]*models.EventMapping, error) {
	filter := bson.M{
		"event_type": eventType,
		"active":     true,
		"$or": []bson.M{
			{"connection_id": connectionID},
			{"connection_id": ""},
			{"connection_id": bson.M{"$exists": false}},
		},
	}

	cursor, err := r.collection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var mappings []*models.EventMapping
	if err := cursor.All(ctx, &mappings); err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *EventMappingRepository) Create(ctx context.Context, mapping *models.EventMapping) error {
	mapping.ID = primitive.NewObjectID().Hex()
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()
	_, err := r.collection().InsertOne(ctx, mapping)
	return err
}

func (r *EventMappingRepository) Update(ctx context.Context, id string, mapping *models.EventMapping) error {
	mapping.UpdatedAt = time.Now()
	_, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": mapping})
	return err
}

func (r *EventMappingRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *EventMappingRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "event_type", Value: 1}, {Key: "connection_id", Value: 1}}},
	})
	return err
}
