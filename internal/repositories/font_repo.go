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

const fontsCollection = "fonts"

type FontRepository struct {
	db *mongo.Database
}

func NewFontRepository(db *mongo.Database) *FontRepository {
	return &FontRepository{db: db}
}

func (r *FontRepository) collection() *mongo.Collection {
	return r.db.Collection(fontsCollection)
}

func (r *FontRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "family", Value: 1}, {Key: "scope", Value: 1}, {Key: "template_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_family_scope_template"),
	})
	return err
}

func (r *FontRepository) Create(ctx context.Context, font *models.Font) error {
	font.ID = primitive.NewObjectID()
	font.CreatedAt = time.Now()
	if font.Scope == "" {
		font.Scope = models.FontScopeGlobal
	}
	_, err := r.collection().InsertOne(ctx, font)
	return err
}

func (r *FontRepository) GetByID(ctx context.Context, id string) (*models.Font, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var font models.Font
	err = r.collection().FindOne(ctx, bson.M{"_id": objID}).Decode(&font)
	if err != nil {
		return nil, err
	}
	return &font, nil
}

// GetByFamily returns the most relevant font for a family name.
// Lookup order: template-scoped (if templateID given) > global. Empty templateID skips template lookup.
func (r *FontRepository) GetByFamily(ctx context.Context, family, templateID string) (*models.Font, error) {
	if templateID != "" {
		var f models.Font
		err := r.collection().FindOne(ctx, bson.M{"family": family, "scope": models.FontScopeTemplate, "template_id": templateID}).Decode(&f)
		if err == nil {
			return &f, nil
		}
	}
	var f models.Font
	err := r.collection().FindOne(ctx, bson.M{"family": family, "scope": models.FontScopeGlobal}).Decode(&f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FontRepository) List(ctx context.Context, scope string, templateID string) ([]*models.Font, error) {
	filter := bson.M{}
	if scope != "" {
		filter["scope"] = scope
	}
	if templateID != "" {
		filter["template_id"] = templateID
	}
	opts := options.Find().SetSort(bson.M{"family": 1})
	cursor, err := r.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var fonts []*models.Font
	if err := cursor.All(ctx, &fonts); err != nil {
		return nil, err
	}
	return fonts, nil
}

func (r *FontRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection().DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
