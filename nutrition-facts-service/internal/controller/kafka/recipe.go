package kafka

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition"
	"io"

	"github.com/segmentio/kafka-go"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition/service"
)

type RecipeWorker struct {
	nutritionService     service.Service
	recipeReader         *kafka.Reader
	nutritionFactsWriter *kafka.Writer
	reqIngredientsWriter *kafka.Writer
	ingredientsReader    *kafka.Reader
}

func NewRecipeWorker(nutritionService service.Service, brokers []string) (Worker, error) {
	recipeReader, err := newReader(brokers, "nutrition-facts-service-recipes", TopicRecipes)
	if err != nil {
		return nil, err
	}
	ingredientsReader, err := newReader(brokers, "nutrition-facts-service-ingredients", TopicIngredients)
	if err != nil {
		return nil, err
	}
	nutritionFactsWriter := newWriter(brokers, TopicNutritionFacts)
	reqIngredientsWriter := newWriter(brokers, TopicIngredientsReq)
	return RecipeWorker{
		nutritionService:     nutritionService,
		recipeReader:         recipeReader,
		nutritionFactsWriter: nutritionFactsWriter,
		reqIngredientsWriter: reqIngredientsWriter,
		ingredientsReader:    ingredientsReader,
	}, nil
}

func (w RecipeWorker) Process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var dto nutrition.RecipeDTO
		err := readDTO(ctx, w.recipeReader, &dto)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return err
			}
			continue
		}
		log.Info().Msgf("got RecipeDTO: %+v", dto)

		if len(dto.ID) > 0 {
			go w.processRecipe(ctx, dto.Recipe)
		}

	}
}

func (w RecipeWorker) Stop() error {
	return closeAll(w.recipeReader, w.nutritionFactsWriter, w.reqIngredientsWriter, w.ingredientsReader)
}

func (w RecipeWorker) processRecipe(ctx context.Context, recipe nutrition.Recipe) {
	ingredients := make(map[string]nutrition.Ingredient, 0)
	// TODO: process asynchronously or add bulk get API to ingredients-service
	for _, ingredient := range recipe.Ingredients {
		dto := nutrition.FindIngredientsDTO{
			ID: ingredient.IngredientID,
		}

		write(w.reqIngredientsWriter, dto.ID, dto)
		log.Info().Msgf("sent FindIngredientsDTO: %+v", dto)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			var ingredientDTO nutrition.IngredientDTO
			err := readDTO(ctx, w.ingredientsReader, &ingredientDTO)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				continue
			}
			log.Info().Msgf("got IngredientDTO: %+v", ingredientDTO)
			if dto.ID != ingredientDTO.ID {
				continue
			}
			if len(ingredientDTO.Error) > 0 {
				log.Error().Msgf("failed to get ingredient: %s", ingredientDTO.Error)
				// TODO:
				return
			}
			ingredients[ingredientDTO.ID] = ingredientDTO.Ingredient
			break
		}

	}

	recipeNutritionsDTO, err := w.nutritionService.CalcRecipeNutritions(recipe, ingredients)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate recipe nutritions")
		return
	}

	write(w.nutritionFactsWriter, recipeNutritionsDTO.RecipeID, recipeNutritionsDTO)
	log.Info().Msgf("sent RecipeNutritionsDTO: %+v", recipeNutritionsDTO)
}
