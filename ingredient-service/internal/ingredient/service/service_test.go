package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage/mongodb"
	"os"
	"testing"
	"time"
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

	t.Run("add ingredient", func(t *testing.T) {
		dto := ingredient.CreateIngredientDTO{
			Name:     "соль",
			BaseUnit: "г",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := serv.Create(ctx, dto)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})
	t.Run("add ingredient and find by id", func(t *testing.T) {
		dto := ingredient.CreateIngredientDTO{
			Name:     "сахар",
			BaseUnit: "г",
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := serv.Create(ctx, dto)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		ingr, err := serv.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, dto.Name, ingr.Name)
	})
	t.Run("add ingredients and search by name", func(t *testing.T) {
		dtos := []ingredient.CreateIngredientDTO{
			{
				Name:     "гороховая мука",
				BaseUnit: "г",
			},
			{
				Name:     "кукурузная мука",
				BaseUnit: "г",
			},
			{
				Name:     "яблоко",
				BaseUnit: "шт",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, dto := range dtos {
			id, err := serv.Create(ctx, dto)
			require.NoError(t, err)
			assert.NotEmpty(t, id)
		}

		ingrs, err := serv.SearchByName(ctx, "мука")
		require.NoError(t, err)
		assert.Equal(t, 2, len(ingrs))
	})
}
