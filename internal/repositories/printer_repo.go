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

const printersCollection = "printers"

type PrinterRepository struct {
	db *mongo.Database
}

func NewPrinterRepository(db *mongo.Database) *PrinterRepository {
	return &PrinterRepository{db: db}
}

func (r *PrinterRepository) collection() *mongo.Collection {
	return r.db.Collection(printersCollection)
}

func (r *PrinterRepository) Create(ctx context.Context, p *models.Printer) error {
	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.IsActive = true
	if p.Status == "" {
		p.Status = models.PrinterStatusIdle
	}
	_, err := r.collection().InsertOne(ctx, p)
	return err
}

func (r *PrinterRepository) Upsert(ctx context.Context, p *models.Printer) error {
	filter := bson.M{"name": p.Name, "tenant_id": p.TenantID}
	update := bson.M{
		"$set": bson.M{
			"cups_name":   p.CupsName,
			"location":    p.Location,
			"paper_sizes": p.PaperSizes,
			"color_modes": p.ColorModes,
			"is_active":   true,
			"updated_at":  time.Now(),
		},
		"$setOnInsert": bson.M{
			"_id":        primitive.NewObjectID(),
			"status":     models.PrinterStatusIdle,
			"created_at": time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection().UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *PrinterRepository) GetByID(ctx context.Context, id string) (*models.Printer, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var p models.Printer
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PrinterRepository) List(ctx context.Context, page, limit int64) ([]*models.Printer, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*models.Printer
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PrinterRepository) ListByTenantID(ctx context.Context, tenantID string, page, limit int64) ([]*models.Printer, error) {
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})
	filter := bson.M{"is_active": true}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}
	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*models.Printer
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PrinterRepository) Update(ctx context.Context, p *models.Printer) error {
	p.UpdatedAt = time.Now()
	_, err := r.collection().ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	return err
}

func (r *PrinterRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
