package mongodb

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	dsn := os.Getenv("TEST_MONGO_DSN")
	if len(dsn) == 0 {
		dsn = "mongodb://dev:dev@localhost:27017/test?authSource=admin"
	}
	db, cleanup := getTestDB(t, dsn)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := cleanup(ctx)
		if err != nil {
			t.Fatalf("test cleanup failed: %s", err)
		}
	}()

	s, err := New(db, dsn)
	if err != nil {
		t.Fatalf("could not initialize storage: %s", err)
	}

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

		got, err := s.FindById(ctx, id)
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

		got, err := s.SearchByName(ctx, "сахар", false)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(got))
	})
}

func getTestDB(t *testing.T, dsn string) (*mongo.Database, func(ctx context.Context) error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dsn))
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
		err := client.Database("test").Drop(ctx)
		log.Err(err)
		return client.Disconnect(ctx)
	}

	return client.Database("test"), cleanup
}
