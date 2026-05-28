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

const printJobsCollection = "print_jobs"

type PrintJobRepository struct {
	db *mongo.Database
}

func NewPrintJobRepository(db *mongo.Database) *PrintJobRepository {
	return &PrintJobRepository{db: db}
}

func (r *PrintJobRepository) collection() *mongo.Collection {
	return r.db.Collection(printJobsCollection)
}

func (r *PrintJobRepository) Create(ctx context.Context, pj *models.PrintJob) error {
	pj.ID = primitive.NewObjectID()
	pj.CreatedAt = time.Now()
	pj.UpdatedAt = time.Now()
	if pj.Status == "" {
		pj.Status = models.PrintJobStatusPending
	}
	_, err := r.collection().InsertOne(ctx, pj)
	return err
}

func (r *PrintJobRepository) GetByID(ctx context.Context, id string) (*models.PrintJob, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var pj models.PrintJob
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&pj)
	if err != nil {
		return nil, err
	}
	return &pj, nil
}

func (r *PrintJobRepository) GetByOrderID(ctx context.Context, orderID string) (*models.PrintJob, error) {
	var pj models.PrintJob
	err := r.collection().FindOne(ctx, bson.M{"order_id": orderID}).Decode(&pj)
	if err != nil {
		return nil, err
	}
	return &pj, nil
}

func (r *PrintJobRepository) List(ctx context.Context, page, limit int64) ([]*models.PrintJob, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*models.PrintJob
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PrintJobRepository) UpdateStatus(ctx context.Context, id string, status models.PrintJobStatus, errorMsg string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	if errorMsg != "" {
		update["$set"].(bson.M)["error_msg"] = errorMsg
	}
	if status == models.PrintJobStatusDone || status == models.PrintJobStatusFailed {
		now := time.Now()
		update["$set"].(bson.M)["completed_at"] = now
	}
	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": objID}, update)
	return err
}

func (r *PrintJobRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *PrintJobRepository) UpdateCUPSJobID(ctx context.Context, id, cupsJobID string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"cups_job_id": cupsJobID,
			"status":      models.PrintJobStatusPrinting,
			"updated_at":   time.Now(),
		},
	})
	return err
}
