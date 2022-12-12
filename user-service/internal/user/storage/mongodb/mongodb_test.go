package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/user-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
)

func TestStorage(t *testing.T) {
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}

	s, cleanup, err := NewTestStorage(dsn, "test")
	if err != nil {
		t.Fatalf("could not initialize test storage: %s", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cleanup(ctx)
		if err != nil {
			t.Fatalf("test cleanup failed: %s", err)
		}
	}()

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
