package kafka

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/nutrition-facts-service/internal/nutrition/service"
	"golang.org/x/sync/errgroup"
	"strings"
)

type kafkaController struct {
	workers []Worker
}

type Worker interface {
	Process(ctx context.Context) error
	Stop() error
}

func NewController(nutritionService service.Service, kafkaBrokerURLs string) (controller.Controller, error) {
	brokers := strings.Split(kafkaBrokerURLs, ",")

	var workers []Worker

	recipeWorker, err := NewRecipeWorker(nutritionService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, recipeWorker)

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
