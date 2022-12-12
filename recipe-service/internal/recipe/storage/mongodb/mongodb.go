package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	apperror "github.com/tony-spark/recipetor-backend/recipe-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewStorage(dsn string, dbname string) (storage.Storage, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, fmt.Errorf("could not create connection to test DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to test DB: %w", err)
	}

	db := client.Database(dbname)

	collection := db.Collection("ingredients")
	return mongoStorage{
		client:     client,
		collection: collection,
	}, nil
}

func NewTestStorage(dsn string, dbname string) (storage.Storage, func(ctx context.Context) error, error) {
	stor, err := NewStorage(dsn, dbname)
	if err != nil {
		return nil, nil, err
	}
	mongoStor := stor.(mongoStorage)
	cleanup := func(ctx context.Context) error {
		err := mongoStor.client.Database(dbname).Drop(ctx)
		log.Err(err)
		return mongoStor.client.Disconnect(ctx)
	}

	return stor, cleanup, nil
}

func (m mongoStorage) Create(ctx context.Context, recipe recipe.Recipe) (string, error) {
	result, err := m.collection.InsertOne(ctx, recipe)
	if err != nil {
		return "", fmt.Errorf("failed to insert recipe: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to insert ingredient: invalid IntertedID")
	}

	return id.Hex(), nil
}

func (m mongoStorage) GetByID(ctx context.Context, id string) (r recipe.Recipe, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return r, fmt.Errorf("wrong id: %w", err)
	}
	result := m.collection.FindOne(ctx, bson.M{"_id": oid})
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return r, apperror.ErrNotFound
		}
		err = result.Err()
		return r, fmt.Errorf("failed to find recipe: %w", err)
	}
	err = result.Decode(&r)
	if err != nil {
		return r, fmt.Errorf("failed to fetch recipe: %w", err)
	}
	return
}

func (m mongoStorage) GetAll(ctx context.Context, start int64, limit int64) (rs []recipe.Recipe, err error) {
	cursor, err := m.collection.Find(ctx, bson.D{{}}, options.Find().SetSkip(start).SetLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to find recipes: %w", err)
	}

	err = cursor.All(ctx, &rs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}

	return
}

func (m mongoStorage) GetAllByUser(ctx context.Context, userID string) (rs []recipe.Recipe, err error) {
	_, err = primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("wrong id: %w", err)
	}
	cursor, err := m.collection.Find(ctx, bson.M{"created_by": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find recipes: %w", err)
	}
	err = cursor.All(ctx, &rs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %w", err)
	}
	return
}

func (m mongoStorage) Update(ctx context.Context, recipe recipe.Recipe) error {
	oid, err := primitive.ObjectIDFromHex(recipe.ID)
	if err != nil {
		return err
	}

	bs, err := bson.Marshal(recipe)
	if err != nil {
		return err
	}

	var updates bson.M
	err = bson.Unmarshal(bs, &updates)
	if err != nil {
		return err
	}
	delete(updates, "_id")

	result, err := m.collection.UpdateByID(ctx, oid, bson.M{"$set": updates})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (m mongoStorage) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("wrong id: %w", err)
	}
	result, err := m.collection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("error while deleting recipe: %w", err)
	}
	if result.DeletedCount == 0 {
		return apperror.ErrNotFound
	}
	return nil
}
