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

type AddRecipeWorker struct {
	recipeService   service.Service
	newRecipeReader *kafka.Reader
	recipesWriter   *kafka.Writer
}

func NewAddRecipeWorker(recipeService service.Service, brokers []string) (Worker, error) {
	newRecipeReader, err := newReader(brokers, "recipe-service-new", TopicRecipesNew)
	if err != nil {
		return nil, err
	}
	recipesWriter := newWriter(brokers, TopicRecipes)
	return AddRecipeWorker{
		recipeService:   recipeService,
		newRecipeReader: newRecipeReader,
		recipesWriter:   recipesWriter,
	}, nil
}

func (w AddRecipeWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto recipe.CreateRecipeDTO
		corID, err := readDTO(ctx, w.newRecipeReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got CreateRecipeDTO: %+v", dto)

		cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
		id, err := w.recipeService.Create(cntx, dto)
		cancel()
		recipeDTO := recipe.RecipeDTO{
			ID: id,
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to add recipe")
			recipeDTO.Error = err.Error()
		} else {
			recipeDTO.Recipe = recipe.Recipe{
				ID:          id,
				Name:        dto.Name,
				CreatedBy:   dto.CreatedBy,
				Ingredients: dto.Ingredients,
				Steps:       dto.Steps,
			}
		}

		write(w.recipesWriter, dto.Name, recipeDTO, corID)
		log.Info().Msgf("sent RecipeDTO: %+v", recipeDTO)
	}

}

func (w AddRecipeWorker) Stop() error {
	return closeAll(w.newRecipeReader, w.recipesWriter)
}
