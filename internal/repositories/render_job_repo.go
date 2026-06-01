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

const renderJobsCollection = "render_jobs"

type RenderJobRepository struct {
	db *mongo.Database
}

func NewRenderJobRepository(db *mongo.Database) *RenderJobRepository {
	return &RenderJobRepository{db: db}
}

func (r *RenderJobRepository) collection() *mongo.Collection {
	return r.db.Collection(renderJobsCollection)
}

func (r *RenderJobRepository) Create(ctx context.Context, job *models.RenderJob) error {
	job.ID = primitive.NewObjectID()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	if job.Status == "" {
		job.Status = models.RenderStatusPending
	}
	_, err := r.collection().InsertOne(ctx, job)
	return err
}

func (r *RenderJobRepository) GetByID(ctx context.Context, id string) (*models.RenderJob, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var job models.RenderJob
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *RenderJobRepository) List(ctx context.Context, page, limit int64) ([]*models.RenderJob, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*models.RenderJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *RenderJobRepository) Update(ctx context.Context, job *models.RenderJob) error {
	job.UpdatedAt = time.Now()
	_, err := r.collection().ReplaceOne(ctx, bson.M{"_id": job.ID}, job)
	return err
}

func (r *RenderJobRepository) UpdateStatus(ctx context.Context, id string, status models.RenderJobStatus, outputImagePath string, errMsg string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"status":     status,
		"updated_at": time.Now(),
	}
	if outputImagePath != "" {
		update["output_image_path"] = outputImagePath
	}
	if errMsg != "" {
		update["error"] = errMsg
	}
	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": update})
	return err
}

func (r *RenderJobRepository) Count(ctx context.Context) (int64, error) {
	return r.collection().CountDocuments(ctx, bson.M{})
}

func (r *RenderJobRepository) CountByStatus(ctx context.Context, status models.RenderJobStatus) (int64, error) {
	return r.collection().CountDocuments(ctx, bson.M{"status": status})
}

func (r *RenderJobRepository) GetByTemplateID(ctx context.Context, templateID primitive.ObjectID, page, limit int64) ([]*models.RenderJob, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection().Find(ctx, bson.M{"template_id": templateID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*models.RenderJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}
