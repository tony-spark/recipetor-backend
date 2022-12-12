package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/storage/mongodb"
)

func TestService(t *testing.T) {
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}

	stor, cleanup, err := mongodb.NewTestStorage(dsn, "test")
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

	serv := NewService(stor)

	t.Run("create user wrong email", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user@test.",
			Password: "1234",
		}
		_, err := serv.Create(ctx, dto)
		assert.Error(t, err)
	})
	t.Run("create user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user1@test.com",
			Password: "12345",
		}
		id, err := serv.Create(ctx, dto)
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
		createdID, err := serv.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdID)

		u, err := serv.GetByEmailAndPassword(ctx, dto.Email, dto.Password)
		assert.NoError(t, err)
		assert.Equal(t, createdID, u.ID)
	})
	t.Run("create user and get by id", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		dto := user.CreateUserDTO{
			Email:    "user3@test.com",
			Password: "12345",
		}
		createdID, err := serv.Create(ctx, dto)
		assert.NoError(t, err)
		assert.NotEmpty(t, createdID)

		u, err := serv.GetByID(ctx, createdID)
		assert.NoError(t, err)
		assert.Equal(t, u.Email, dto.Email)
	})
}
