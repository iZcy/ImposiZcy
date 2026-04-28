package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const internalRolesCollection = "internal_roles"

type InternalRoleRepository struct {
	db *mongo.Database
}

func NewInternalRoleRepository(db *mongo.Database) *InternalRoleRepository {
	return &InternalRoleRepository{db: db}
}

func (r *InternalRoleRepository) collection() *mongo.Collection {
	return r.db.Collection(internalRolesCollection)
}

func (r *InternalRoleRepository) Create(ctx context.Context, role *models.InternalRole) error {
	role.ID = primitive.NewObjectID().Hex()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	_, err := r.collection().InsertOne(ctx, role)
	return err
}

func (r *InternalRoleRepository) GetAll(ctx context.Context) ([]*models.InternalRole, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.InternalRole
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *InternalRoleRepository) Update(ctx context.Context, id string, role string, description string) error {
	_, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"role":        role,
		"description": description,
		"updated_at":  time.Now(),
	}})
	return err
}

func (r *InternalRoleRepository) List(ctx context.Context) ([]*models.InternalRole, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.InternalRole
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *InternalRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}
