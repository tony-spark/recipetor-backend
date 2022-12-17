package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/storage/mongodb"
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

	t.Run("create recipe", func(t *testing.T) {
		dto := recipe.CreateRecipeDTO{
			Name:        "Рецепт 1",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := serv.Create(ctx, dto)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("create recipe, get by id and update", func(t *testing.T) {
		dto := recipe.CreateRecipeDTO{
			Name:        "Рецепт 3",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := serv.Create(ctx, dto)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		got, err := serv.GetByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, dto.Name, got.Name)

		updateDTO := recipe.UpdateRecipeDTO{
			ID:          got.ID,
			Name:        "Рецепт 3 (ред.)",
			Ingredients: got.Ingredients,
			Steps:       got.Steps,
		}
		err = serv.Update(ctx, updateDTO)
		require.NoError(t, err)
	})
}
