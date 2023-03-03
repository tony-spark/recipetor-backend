package kafka

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/recipe/service"
	"golang.org/x/sync/errgroup"
)

type kafkaController struct {
	workers []Worker
}

type Worker interface {
	Process(ctx context.Context) error
	Stop() error
}

func NewController(recipeService service.Service, kafkaBrokerURLs string) (controller.Controller, error) {
	brokers := strings.Split(kafkaBrokerURLs, ",")

	var workers []Worker

	newRecipeWorker, err := NewAddRecipeWorker(recipeService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, newRecipeWorker)

	recipeNutritionFactsWorker, err := NewRecipeNutritionFactsWorker(recipeService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, recipeNutritionFactsWorker)

	findRecipesWorker, err := NewFindRecipesWorker(recipeService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, findRecipesWorker)

	return kafkaController{
		workers: workers,
	}, nil
}

func (k kafkaController) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	for _, w := range k.workers {
		worker := w
		group.Go(func() error {
			return worker.Process(ctx)
		})
	}
	return group.Wait()
}

func (k kafkaController) Stop() error {
	var result error
	for _, w := range k.workers {
		err := w.Stop()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}
