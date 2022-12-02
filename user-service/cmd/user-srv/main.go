package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tony-spark/recipetor-backend/user-service/internal/model"
)

func main() {
	fmt.Println("starting user service...")

	var u model.User
	fmt.Printf("- %+v\n", u)

	terminateSignal := make(chan os.Signal, 1)
	signal.Notify(terminateSignal, syscall.SIGINT, syscall.SIGTERM)

	<-terminateSignal
	fmt.Println("user service interrupted via system signal")
}
