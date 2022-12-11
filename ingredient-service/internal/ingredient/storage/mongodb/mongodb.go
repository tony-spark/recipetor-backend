package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	edb "github.com/tony-spark/recipetor-backend/ingredient-service/db"
	apperror "github.com/tony-spark/recipetor-backend/ingredient-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoStorage struct {
	collection *mongo.Collection
}

func New(db *mongo.Database, dsn string) (storage.Storage, error) {
	driver, err := iofs.New(edb.EmbeddedDBFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("could not open migrations: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not initialize migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("could not execute db migrations: %w", err)
	}

	collection := db.Collection("ingredients")
	return mongoStorage{
		collection: collection,
	}, nil
}

func (m mongoStorage) Create(ctx context.Context, ingredient ingredient.Ingredient) (string, error) {
	result, err := m.collection.InsertOne(ctx, ingredient)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to insert ingredient: invalid IntertedID")
	}

	return id.Hex(), nil
}

func (m mongoStorage) FindById(ctx context.Context, id string) (ing ingredient.Ingredient, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	result := m.collection.FindOne(ctx, bson.M{"_id": oid})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return ing, apperror.ErrNotFound
		}
		err = result.Err()
		return
	}
	err = result.Decode(&ing)
	return
}

func (m mongoStorage) SearchByName(ctx context.Context, query string, fillNutritionInfo bool) (ings []ingredient.Ingredient, err error) {
	filter := bson.D{{"$text", bson.D{{"$search", query}}}}

	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return
	}

	err = cursor.All(ctx, &ings)
	if err != nil {
		return
	}

	return
}
