package service

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	stor, cleanup := getTestStorage(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cleanup(ctx)
		if err != nil {
			t.Fatalf("test cleanup failed: %s", err)
		}
	}()

	s := NewService(stor)

	t.Run("create user wrong email", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user@test.",
			Password: "1234",
		}
		_, err := s.Create(ctx, dto)
		assert.Error(t, err)
	})
	t.Run("create user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user1@test.com",
			Password: "12345",
		}
		id, err := s.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
	})
	t.Run("create user and get by email password", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user2@test.com",
			Password: "12345",
		}
		createdId, err := s.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdId)

		u, err := s.GetByEmailAndPassword(ctx, dto.Email, dto.Password)
		assert.NoError(t, err)
		assert.Equal(t, createdId, u.Id)
	})
	t.Run("create user and get by id", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user3@test.com",
			Password: "12345",
		}
		createdId, err := s.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdId)

		u, err := s.GetById(ctx, createdId)
		assert.NoError(t, err)
		assert.Equal(t, u.Email, dto.Email)
	})
}

func getTestStorage(t *testing.T) (storage.Storage, func(ctx context.Context) error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://dev:dev@localhost:27017"))
	if err != nil {
		t.Fatalf("could not create connection to test DB: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		t.Fatalf("could not connect to test DB: %s", err)
	}
	cleanup := func(ctx context.Context) error {
		err := client.Database("test").Collection("users").Drop(ctx)
		log.Err(err)
		return client.Disconnect(ctx)
	}

	db := client.Database("test")

	return mongodb.NewStorage(db), cleanup
}
