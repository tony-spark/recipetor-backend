package kafka

import (
	"context"
	"golang.org/x/sync/errgroup"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/tony-spark/recipetor-backend/user-service/internal/controller"
	"github.com/tony-spark/recipetor-backend/user-service/internal/user/service"
)

type kafkaController struct {
	workers []Worker
}

type Worker interface {
	Process(ctx context.Context) error
	Stop() error
}

func NewController(userService service.Service, kafkaBrokerURLs string) (controller.Controller, error) {
	brokers := strings.Split(kafkaBrokerURLs, ",")

	var workers []Worker
	regWorker, err := NewRegistrationWorker(userService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, regWorker)

	loginWorker, err := NewLoginWorker(userService, brokers)
	if err != nil {
		return nil, err
	}
	workers = append(workers, loginWorker)

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
