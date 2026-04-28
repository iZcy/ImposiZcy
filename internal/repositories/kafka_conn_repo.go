package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const kafkaConnectionsCollection = "kafka_connections"

type KafkaConnectionRepository struct {
	db *mongo.Database
}

func NewKafkaConnectionRepository(db *mongo.Database) *KafkaConnectionRepository {
	return &KafkaConnectionRepository{db: db}
}

func (r *KafkaConnectionRepository) collection() *mongo.Collection {
	return r.db.Collection(kafkaConnectionsCollection)
}

type KafkaConnection struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Brokers   []string           `bson:"brokers" json:"brokers"`
	Topics    []string           `bson:"topics" json:"topics"`
	GroupID   string             `bson:"group_id" json:"group_id"`
	Active    bool               `bson:"active" json:"active"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func (r *KafkaConnectionRepository) Create(ctx context.Context, conn *KafkaConnection) error {
	conn.ID = primitive.NewObjectID()
	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()
	_, err := r.collection().InsertOne(ctx, conn)
	return err
}

func (r *KafkaConnectionRepository) GetByID(ctx context.Context, id string) (*KafkaConnection, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var conn KafkaConnection
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *KafkaConnectionRepository) List(ctx context.Context) ([]*KafkaConnection, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conns []*KafkaConnection
	if err := cursor.All(ctx, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func (r *KafkaConnectionRepository) Update(ctx context.Context, id string, update bson.M) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["updated_at"] = time.Now()
	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

func (r *KafkaConnectionRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *KafkaConnectionRepository) ListActive(ctx context.Context) ([]*KafkaConnection, error) {
	cursor, err := r.collection().Find(ctx, bson.M{"active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conns []*KafkaConnection
	if err := cursor.All(ctx, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func (r *KafkaConnectionRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
	})
	return err
}
