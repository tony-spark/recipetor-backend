package controller

type Controller interface {
	Run() error
	Stop() error
}
