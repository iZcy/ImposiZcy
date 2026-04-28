package repositories

import (
	"context"
	"time"

	"github.com/iZcy/imposizcy/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const externalRolesCollection = "external_roles"

type ExternalRoleRepository struct {
	db *mongo.Database
}

func NewExternalRoleRepository(db *mongo.Database) *ExternalRoleRepository {
	return &ExternalRoleRepository{db: db}
}

func (r *ExternalRoleRepository) collection() *mongo.Collection {
	return r.db.Collection(externalRolesCollection)
}

func (r *ExternalRoleRepository) Create(ctx context.Context, role *models.ExternalRole) error {
	role.ID = primitive.NewObjectID().Hex()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	_, err := r.collection().InsertOne(ctx, role)
	return err
}

func (r *ExternalRoleRepository) GetAll(ctx context.Context) ([]*models.ExternalRole, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.ExternalRole
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *ExternalRoleRepository) GetByID(ctx context.Context, id string) (*models.ExternalRole, error) {
	var role models.ExternalRole
	err := r.collection().FindOne(ctx, bson.M{"_id": id}).Decode(&role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *ExternalRoleRepository) Update(ctx context.Context, id string, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *ExternalRoleRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection().DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ExternalRoleRepository) List(ctx context.Context) ([]*models.ExternalRole, error) {
	cursor, err := r.collection().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []*models.ExternalRole
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}
