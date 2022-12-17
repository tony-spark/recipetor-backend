package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	edb "github.com/tony-spark/recipetor-backend/ingredient-service/db"
	apperror "github.com/tony-spark/recipetor-backend/ingredient-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type mongoStorage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewStorage(dsn string, dbname string) (storage.Storage, error) {
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

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
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

func (m mongoStorage) Create(ctx context.Context, ingredient ingredient.Ingredient) (string, error) {
	result, err := m.collection.InsertOne(ctx, ingredient)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", apperror.ErrDuplicate
		}
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("failed to insert ingredient: invalid IntertedID")
	}

	return id.Hex(), nil
}

func (m mongoStorage) FindByID(ctx context.Context, id string) (ing ingredient.Ingredient, err error) {
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

func (m mongoStorage) SearchByName(ctx context.Context, query string) (ings []ingredient.Ingredient, err error) {
	filter := bson.D{{Key: "$text",
		Value: bson.D{{Key: "$search", Value: query}}}}

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
