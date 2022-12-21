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

type FindRecipesWorker struct {
	recipeService    service.Service
	reqRecipesReader *kafka.Reader
	recipeWriter     *kafka.Writer
}

func NewFindRecipesWorker(recipeService service.Service, brokers []string) (Worker, error) {
	reqRecipesReader, err := newReader(brokers, "recipe-service-find", TopicRecipesReq)
	if err != nil {
		return nil, err
	}
	recipesWriter := newWriter(brokers, TopicRecipes)
	return FindRecipesWorker{
		recipeService:    recipeService,
		reqRecipesReader: reqRecipesReader,
		recipeWriter:     recipesWriter,
	}, nil
}

func (w FindRecipesWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto recipe.FindRecipeDTO
		err := readDTO(ctx, w.reqRecipesReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got FindRecipeDTO: %+v", dto)

		if len(dto.ID) > 0 {
			cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
			recip, err := w.recipeService.GetByID(cntx, dto.ID)
			cancel()

			recipeDTO := recipe.RecipeDTO{
				ID: dto.ID,
			}
			if err != nil {
				log.Error().Err(err).Msg("failed to find recipe")
			} else {
				recipeDTO.Recipe = recip
			}

			write(w.recipeWriter, dto.ID, recipeDTO)
			log.Info().Msgf("sent RecipeDTO: %+v", recipeDTO)
		}

		if len(dto.UserID) > 0 {
			cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
			recipes, err := w.recipeService.GetAllByUser(cntx, dto.UserID)
			cancel()

			if err != nil {
				log.Error().Err(err).Msg("failed to find recipes")
				write(w.recipeWriter, dto.UserID, recipe.RecipeDTO{
					UserID: dto.UserID,
					Error:  err.Error(),
				})
			} else {
				for _, recip := range recipes {
					write(w.recipeWriter, dto.UserID, recipe.RecipeDTO{
						Recipe: recip,
						UserID: dto.UserID,
					})
				}
			}
		}

		if len(dto.IngredientIDs) > 0 {
			// TODO: implement
		}
	}
}

func (w FindRecipesWorker) Stop() error {
	return closeAll(w.reqRecipesReader, w.recipeWriter)
}
