package services

import (
	"context"
	"fmt"
	"time"

	"github.com/iZcy/imposizcy/config"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoService struct {
	client   *mongo.Client
	database *mongo.Database
	logger   *logrus.Logger
}

func NewMongoService(cfg *config.Config, logger *logrus.Logger) (*MongoService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoDB.URI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.MongoDB.Database)

	logger.WithFields(logrus.Fields{
		"uri":      cfg.MongoDB.URI,
		"database": cfg.MongoDB.Database,
	}).Info("Connected to MongoDB successfully")

	return &MongoService{
		client:   client,
		database: database,
		logger:   logger,
	}, nil
}

func (s *MongoService) GetDatabase() *mongo.Database {
	return s.database
}

func (s *MongoService) GetCollection(name string) *mongo.Collection {
	return s.database.Collection(name)
}

func (s *MongoService) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
