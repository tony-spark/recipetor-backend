package mongodb

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperror "github.com/tony-spark/recipetor-backend/recipe-service/internal/errors"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
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

	t.Run("create recipe", func(t *testing.T) {
		r := recipe.Recipe{
			Name:        "Тестовый рецепт 1",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, r)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("create recipe and get by id", func(t *testing.T) {
		r := recipe.Recipe{
			Name:        "Тестовый рецепт 2",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, r)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		got, err := s.GetById(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, r.Name, got.Name)
	})

	t.Run("create recipe and get all", func(t *testing.T) {
		loaded := []recipe.Recipe{
			{
				Name:        "Тестовый рецепт 3",
				CreatedBy:   "639673eb2c5bcae361a8ad4a",
				Ingredients: nil,
				Steps:       nil,
			},
			{
				Name:        "Тестовый рецепт 4",
				CreatedBy:   "639673eb2c5bcae361a8ad4a",
				Ingredients: nil,
				Steps:       nil,
			},
			{
				Name:        "Тестовый рецепт 5",
				CreatedBy:   "639673eb2c5bcae361a8ad4a",
				Ingredients: nil,
				Steps:       nil,
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, r := range loaded {
			id, err := s.Create(ctx, r)
			require.NoError(t, err)
			assert.NotEmpty(t, id)
		}

		{
			rs, err := s.GetAll(ctx, 0, 2)
			require.NoError(t, err)
			assert.Condition(t, func() (success bool) {
				return len(rs) <= 2
			}, "results count should be >= 2")
		}
	})

	t.Run("create recipe and get all by user", func(t *testing.T) {
		loaded := []recipe.Recipe{
			{
				Name:        "Тестовый рецепт 6",
				CreatedBy:   "639673eb2c5bcae361a8ad4f",
				Ingredients: nil,
				Steps:       nil,
			},
			{
				Name:        "Тестовый рецепт 7",
				CreatedBy:   "639673eb2c5bcae361a8ad4f",
				Ingredients: nil,
				Steps:       nil,
			},
			{
				Name:        "Тестовый рецепт 8",
				CreatedBy:   "639673eb2c5bcae361a8ad4a",
				Ingredients: nil,
				Steps:       nil,
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, r := range loaded {
			id, err := s.Create(ctx, r)
			require.NoError(t, err)
			assert.NotEmpty(t, id)
		}

		{
			rs, err := s.GetAllByUser(ctx, "639673eb2c5bcae361a8ad4f")
			require.NoError(t, err)
			assert.Equal(t, 2, len(rs))
		}
	})

	t.Run("create recipe and update", func(t *testing.T) {
		r := recipe.Recipe{
			Name:        "Тестовый рецепт 9",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, r)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		r.ID = id
		updated := r
		updated.Name = "Тестовый рецепт 9 (ред.)"

		err = s.Update(ctx, updated)
		require.NoError(t, err)

		got, err := s.GetById(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, updated.Name, got.Name)
	})

	t.Run("create recipe and delete", func(t *testing.T) {
		r := recipe.Recipe{
			Name:        "Тестовый рецепт 10",
			CreatedBy:   "639673eb2c5bcae361a8ad4a",
			Ingredients: nil,
			Steps:       nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, r)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		err = s.Delete(ctx, id)
		require.NoError(t, err)

		got, err := s.GetById(ctx, id)
		assert.EqualError(t, err, apperror.ErrNotFound.Error())
		assert.Empty(t, got)
	})
}
