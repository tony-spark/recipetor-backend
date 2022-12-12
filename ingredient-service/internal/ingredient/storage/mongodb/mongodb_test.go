package mongodb

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/storage"
	"os"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}

	s, cleanup, err := getTestStorage(dsn, "test")
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

	t.Run("create ingredient", func(t *testing.T) {
		i := ingredient.Ingredient{
			Name:     "мука",
			BaseUnit: "г",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, i)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})
	t.Run("create ingredient and get by id", func(t *testing.T) {
		inserted := ingredient.Ingredient{
			Name:     "перец молотый",
			BaseUnit: "г",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := s.Create(ctx, inserted)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		got, err := s.FindByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, inserted.Name, got.Name)
	})
	t.Run("create ingredients and search by name", func(t *testing.T) {
		ingredients := []ingredient.Ingredient{
			{
				Name:     "сахар",
				BaseUnit: "г",
			},
			{
				Name:     "ванильный сахар",
				BaseUnit: "г",
			},
			{
				Name:     "соль",
				BaseUnit: "г",
			},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, ing := range ingredients {
			id, err := s.Create(ctx, ing)
			require.NoError(t, err)
			assert.NotEmpty(t, id)
		}

		got, err := s.SearchByName(ctx, "сахар")
		assert.NoError(t, err)
		assert.Equal(t, 2, len(got))
	})
}

func getTestStorage(dsn string, dbname string) (storage.Storage, func(ctx context.Context) error, error) {
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
