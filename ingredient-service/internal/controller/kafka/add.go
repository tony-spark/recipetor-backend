package kafka

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/service"
	"io"
	"time"
)

type AddIngredientWorker struct {
	ingredientService    service.Service
	newIngredientsReader *kafka.Reader
	ingredientsWriter    *kafka.Writer
}

func NewAddIngredientWorker(ingredientService service.Service, brokers []string) (Worker, error) {
	newIngredientsReader, err := newReader(brokers, "ingredients-service-new", TopicIngredientsNew)
	if err != nil {
		return nil, err
	}
	ingredientsWriter := newWriter(brokers, TopicIngredients)
	return AddIngredientWorker{
		ingredientService:    ingredientService,
		newIngredientsReader: newIngredientsReader,
		ingredientsWriter:    ingredientsWriter,
	}, nil
}

func (w AddIngredientWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto ingredient.CreateIngredientDTO
		err := readDTO(ctx, w.newIngredientsReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got CreateIngredientDTO: %+v", dto)

		cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
		id, err := w.ingredientService.Create(cntx, dto)
		ingredientDTO := ingredient.IngredientDTO{
			Name: dto.Name,
		}
		cancel()
		if err != nil {
			log.Error().Err(err).Msg("failed to add ingredient")
			ingredientDTO.Error = err.Error()
		} else {
			ingredientDTO.Ingredient = ingredient.Ingredient{
				ID:             id,
				Name:           dto.Name,
				BaseUnit:       dto.BaseUnit,
				NutritionFacts: dto.NutritionFacts,
			}
		}

		write(w.ingredientsWriter, dto.Name, ingredientDTO)
		log.Info().Msgf("sent IngredientDTO: %+v", ingredientDTO)
	}
}

func (w AddIngredientWorker) Stop() error {
	return closeAll(w.newIngredientsReader, w.ingredientsWriter)
}
