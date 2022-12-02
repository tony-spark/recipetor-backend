package main

import (
	"fmt"
	"github.com/tony-spark/recipetor-backend/ingredient-service/internal/model"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("starting ingredient service...")

	var i model.Ingredient
	fmt.Printf("%+v\n", i)

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	fmt.Println("ingredient service interrupted via system signal")
}
