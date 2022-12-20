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

type FindIngredientsWorker struct {
	ingredientService    service.Service
	ingredientsReqReader *kafka.Reader
	ingredientsWriter    *kafka.Writer
}

func NewFindIngredientsWorker(ingredientService service.Service, brokers []string) (Worker, error) {
	ingredientsReqReader, err := newReader(brokers, "ingredients-service-find", TopicIngredientsReq)
	if err != nil {
		return nil, err
	}
	ingredientsWriter := newWriter(brokers, TopicIngredients)
	return FindIngredientsWorker{
		ingredientService:    ingredientService,
		ingredientsReqReader: ingredientsReqReader,
		ingredientsWriter:    ingredientsWriter,
	}, nil
}

func (w FindIngredientsWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto ingredient.FindIngredientsDTO
		err := readDTO(ctx, w.ingredientsReqReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got FindIngredientsDTO: %+v", dto)

		if len(dto.ID) > 0 {
			cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
			ingr, err := w.ingredientService.GetByID(cntx, dto.ID)
			cancel()

			ingredientDTO := ingredient.IngredientDTO{
				ID: dto.ID,
			}
			if err != nil {
				log.Error().Err(err).Msg("failed to find ingredient")
				ingredientDTO.Error = err.Error()
			} else {
				ingredientDTO.Ingredient = ingr
			}

			write(w.ingredientsWriter, dto.ID, ingredientDTO)
			log.Info().Msgf("sent IngredientDTO: %+v", ingredientDTO)
		}

		if len(dto.NameQuery) > 0 {
			cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
			ingredients, err := w.ingredientService.SearchByName(cntx, dto.NameQuery)
			cancel()

			if err != nil {
				log.Error().Err(err).Msg("failed to find ingredients")
				ingredientDTO := ingredient.IngredientDTO{
					Error:     err.Error(),
					NameQuery: dto.NameQuery,
				}
				write(w.ingredientsWriter, dto.NameQuery, ingredientDTO)
				log.Info().Msgf("sent IngredientDTO: %+v", ingredientDTO)
			}

			for _, ingr := range ingredients {
				ingredientDTO := ingredient.IngredientDTO{
					Ingredient: ingr,
					NameQuery:  dto.NameQuery,
				}
				write(w.ingredientsWriter, dto.NameQuery, ingredientDTO)
				log.Info().Msgf("sent IngredientDTO: %+v", ingredientDTO)
			}

		}
	}
}

func (w FindIngredientsWorker) Stop() error {
	return closeAll(w.ingredientsReqReader, w.ingredientsWriter)
}
