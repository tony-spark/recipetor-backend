package main

import (
	"fmt"
	"github.com/tony-spark/recipetor-backend/recipe-service/internal/model"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("starting recipe service...")

	var r model.Recipe
	fmt.Printf("%+v\n", r)

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	fmt.Println("recipe service interrupted via system signal")
}
