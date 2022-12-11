package mongodb

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/user-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	db, cleanup := getTestCollection(t)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cleanup(ctx)
		if err != nil {
			t.Fatalf("test cleanup failed: %s", err)
		}
	}()

	s := NewStorage(db)

	t.Run("create user", func(t *testing.T) {
		u := user.User{
			Email:        "test@test.com",
			Password:     "",
			RegisteredAt: time.Now(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, u)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})
	t.Run("create user and get by id", func(t *testing.T) {
		u := user.User{
			Email:        "test1@test.com",
			Password:     "",
			RegisteredAt: time.Now(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, u)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		u, err = s.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, u.Email, "test1@test.com")
	})
	t.Run("get by id not found", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := s.FindByID(ctx, "639361be532c9301e02ff4c0")
		assert.EqualError(t, err, errors.ErrNotFound.Error())
	})
	t.Run("create user and find by email", func(t *testing.T) {
		u := user.User{
			Email:        "test2@test.com",
			Password:     "",
			RegisteredAt: time.Now(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, u)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		u, err = s.FindByEmail(ctx, "test2@test.com")
		require.NoError(t, err)
		assert.Equal(t, u.Email, "test2@test.com")
	})
	t.Run("find by email not found", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := s.FindByEmail(ctx, "notfound@test.com")
		assert.EqualError(t, err, errors.ErrNotFound.Error())
	})
}

func getTestCollection(t *testing.T) (*mongo.Database, func(ctx context.Context) error) {
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

	return client.Database("test"), cleanup
}
