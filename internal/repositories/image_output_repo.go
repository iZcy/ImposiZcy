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

const imageOutputsCollection = "image_outputs"

type ImageOutputRepository struct {
	db *mongo.Database
}

func NewImageOutputRepository(db *mongo.Database) *ImageOutputRepository {
	return &ImageOutputRepository{db: db}
}

func (r *ImageOutputRepository) collection() *mongo.Collection {
	return r.db.Collection(imageOutputsCollection)
}

func (r *ImageOutputRepository) Create(ctx context.Context, output *models.ImageOutput) error {
	if output.ID.IsZero() {
		output.ID = primitive.NewObjectID()
	}
	if output.CreatedAt.IsZero() {
		output.CreatedAt = time.Now()
	}
	_, err := r.collection().InsertOne(ctx, output)
	return err
}

func (r *ImageOutputRepository) GetByID(ctx context.Context, id string) (*models.ImageOutput, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var output models.ImageOutput
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (r *ImageOutputRepository) GetByJobID(ctx context.Context, jobID string) (*models.ImageOutput, error) {
	var output models.ImageOutput
	err := r.collection().FindOne(ctx, bson.M{"render_job_id": jobID}).Decode(&output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (r *ImageOutputRepository) ListByJobID(ctx context.Context, jobID string) ([]*models.ImageOutput, error) {
	cursor, err := r.collection().Find(ctx, bson.M{"render_job_id": jobID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var outputs []*models.ImageOutput
	if err := cursor.All(ctx, &outputs); err != nil {
		return nil, err
	}
	return outputs, nil
}

func (r *ImageOutputRepository) List(ctx context.Context, page, limit int64) ([]*models.ImageOutput, int64, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(skip).SetLimit(limit)

	total, err := r.collection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.collection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var outputs []*models.ImageOutput
	if err := cursor.All(ctx, &outputs); err != nil {
		return nil, 0, err
	}
	return outputs, total, nil
}

func (r *ImageOutputRepository) ListByTemplateID(ctx context.Context, templateID string, page, limit int64) ([]*models.ImageOutput, int64, error) {
	filter := bson.M{"template_id": templateID}
	skip := (page - 1) * limit
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(skip).SetLimit(limit)

	total, err := r.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var outputs []*models.ImageOutput
	if err := cursor.All(ctx, &outputs); err != nil {
		return nil, 0, err
	}
	return outputs, total, nil
}

func (r *ImageOutputRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *ImageOutputRepository) Update(ctx context.Context, output *models.ImageOutput) error {
	_, err := r.collection().ReplaceOne(ctx, bson.M{"_id": output.ID}, output)
	return err
}
