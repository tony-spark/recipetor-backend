package mongodb

import (
	"context"
	"errors"
	"fmt"
	apperror "github.com/tony-spark/recipetor-backend/user-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoStorage struct {
	collection *mongo.Collection
}

func NewStorage(db *mongo.Database) storage.Storage {
	return mongoStorage{
		collection: db.Collection("users"),
	}
}

func (m mongoStorage) Create(ctx context.Context, user user.User) (string, error) {
	result, err := m.collection.InsertOne(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to insert user: invalid InsertedID")
	}

	return id.Hex(), nil
}

func (m mongoStorage) FindByID(ctx context.Context, id string) (user user.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	result := m.collection.FindOne(ctx, bson.M{"_id": oid})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return user, apperror.ErrNotFound
		}
		err = result.Err()
		return
	}
	err = result.Decode(&user)
	return
}

func (m mongoStorage) FindByEmail(ctx context.Context, email string) (user user.User, err error) {
	result := m.collection.FindOne(ctx, bson.M{"email": email})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return user, apperror.ErrNotFound
		}
		err = result.Err()
		return
	}
	err = result.Decode(&user)
	return
}
