package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/romanWienicke/go-app-test/rest"
)

func main() {
	fmt.Println("Starting the application...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Start the REST server in a goroutine
	go func() {
		if err := rest.NewServer(port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Create a channel to listen for interrupt or terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down gracefully...")
}
