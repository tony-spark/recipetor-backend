package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo/options"

	apperror "github.com/tony-spark/recipetor-backend/user-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	db := client.Database(dbname)

	collection := db.Collection("users")
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
