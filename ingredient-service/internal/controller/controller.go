package controller

import "context"

type Controller interface {
	Run(ctx context.Context) error
	Stop() error
}
