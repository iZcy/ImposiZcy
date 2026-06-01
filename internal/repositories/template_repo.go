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

const templatesCollection = "templates"

type TemplateRepository struct {
	db *mongo.Database
}

func NewTemplateRepository(db *mongo.Database) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) collection() *mongo.Collection {
	return r.db.Collection(templatesCollection)
}

func (r *TemplateRepository) Create(ctx context.Context, template *models.PrintTemplate) error {
	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	if template.Status == "" {
		template.Status = models.TemplateStatusDraft
	}
	_, err := r.collection().InsertOne(ctx, template)
	return err
}

func (r *TemplateRepository) GetByID(ctx context.Context, id string) (*models.PrintTemplate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var template models.PrintTemplate
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) GetBySlug(ctx context.Context, slug string) (*models.PrintTemplate, error) {
	var template models.PrintTemplate
	err := r.collection().FindOne(ctx, bson.M{"slug": slug}).Decode(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) List(ctx context.Context, page, limit int64, tag string) ([]*models.PrintTemplate, error) {
	filter := bson.M{"is_active": true}
	if tag != "" {
		filter["tags.name"] = tag
	}

	skip := (page - 1) * limit
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*models.PrintTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *TemplateRepository) Update(ctx context.Context, template *models.PrintTemplate) error {
	template.UpdatedAt = time.Now()
	_, err := r.collection().ReplaceOne(ctx, bson.M{"_id": template.ID}, template)
	return err
}

func (r *TemplateRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *TemplateRepository) Count(ctx context.Context) (int64, error) {
	return r.collection().CountDocuments(ctx, bson.M{})
}
