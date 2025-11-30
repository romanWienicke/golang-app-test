package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/romanWienicke/go-app-test/app"
)

func main() {
	fmt.Println("Starting the application...")

	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found (continuing)")
	}

	app, err := app.NewApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		return
	}

	// Start the REST server in a goroutine
	go func() {
		app.Run()
	}()
}
