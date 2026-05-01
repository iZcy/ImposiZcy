package repositories

import (
	"context"
	"time"

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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name"`
	Broker      string             `bson:"broker" json:"broker"`
	Topic       string             `bson:"topic" json:"topic"`
	GroupID     string             `bson:"group_id" json:"group_id"`
	ClientID    string             `bson:"client_id" json:"client_id"`
	AutoOffset  string             `bson:"auto_offset" json:"auto_offset"`
	Enabled     bool               `bson:"enabled" json:"enabled"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type KafkaConnectionStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Enabled   bool   `json:"enabled"`
}

func NewKafkaConnection(conn *KafkaConnection) *KafkaConnection {
	if conn.GroupID == "" {
		conn.GroupID = "imposizcy-consumer-group"
	}
	if conn.ClientID == "" {
		conn.ClientID = "imposizcy"
	}
	if conn.AutoOffset == "" {
		conn.AutoOffset = "earliest"
	}
	return conn
}

func (r *KafkaConnectionRepository) Create(ctx context.Context, conn *KafkaConnection) error {
	if conn.GroupID == "" {
		conn.GroupID = "imposizcy-consumer-group"
	}
	if conn.ClientID == "" {
		conn.ClientID = "imposizcy"
	}
	if conn.AutoOffset == "" {
		conn.AutoOffset = "earliest"
	}
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

func (r *KafkaConnectionRepository) GetEnabled(ctx context.Context) ([]*KafkaConnection, error) {
	cursor, err := r.collection().Find(ctx, bson.M{"enabled": true})
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

func (r *KafkaConnectionRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
	})
	return err
}
