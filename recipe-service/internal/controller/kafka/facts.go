package kafka

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/service"
)

type RecipeNutritionFactsWorker struct {
	recipeService        service.Service
	nutritionFactsReader *kafka.Reader
}

func NewRecipeNutritionFactsWorker(recipeService service.Service, brokers []string) (Worker, error) {
	nutritionFactsReader, err := newReader(brokers, "recipe-service-nutrition-facts", TopicNutritionFacts)
	if err != nil {
		return nil, err
	}
	return RecipeNutritionFactsWorker{
		recipeService:        recipeService,
		nutritionFactsReader: nutritionFactsReader,
	}, nil
}

func (w RecipeNutritionFactsWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto recipe.RecipeNutritionsDTO
		err := readDTO(ctx, w.nutritionFactsReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got RecipeNutritionsDTO: %+v", dto)

		cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
		recip, err := w.recipeService.GetByID(cntx, dto.RecipeID)
		if err != nil {
			cancel()
			log.Error().Err(err).Msg("could not find recipe to update nutrition facts")
			continue
		}

		recip.NutritionFacts = &dto.NutritionFacts

		err = w.recipeService.Update(cntx, recipe.UpdateRecipeDTO{
			ID:             recip.ID,
			Name:           recip.Name,
			Ingredients:    recip.Ingredients,
			Steps:          recip.Steps,
			NutritionFacts: recip.NutritionFacts,
		})
		if err != nil {
			cancel()
			log.Error().Err(err).Msg("could not update recipe's nutrition facts")
			continue
		}

		cancel()
	}
}

func (w RecipeNutritionFactsWorker) Stop() error {
	return closeAll(w.nutritionFactsReader)
}
