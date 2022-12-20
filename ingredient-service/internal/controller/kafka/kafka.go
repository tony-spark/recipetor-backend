package kafka

import (
	"context"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/ingredient/service"
	"golang.org/x/sync/errgroup"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type kafkaController struct {
	workers []Worker
}

type Worker interface {
	Process(ctx context.Context) error
	Stop() error
}

func NewController(ingredientService service.Service, kafkaBrokerURLs string) (controller.Controller, error) {
	brokers := strings.Split(kafkaBrokerURLs, ",")

	var workers []Worker
	addIngredientWorker, err := NewAddIngredientWorker(ingredientService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, addIngredientWorker)

	findIngredientsWorker, err := NewFindIngredientsWorker(ingredientService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, findIngredientsWorker)

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
