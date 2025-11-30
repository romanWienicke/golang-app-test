package app

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/romanWienicke/go-app-test/foundation/postgres"
	"github.com/romanWienicke/go-app-test/rest"
)

type app struct {
	db *postgres.Db
}

func NewApp() (*app, error) {
	app := &app{}
	if err := app.init(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *app) Run() {
	a.runServer()
}

func (a *app) runServer() {
	port := os.Getenv("HTTP_PORT")
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

	log.Println("Shutting down gracefully...")
	if err := a.Close(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}

func (a *app) Close() error {
	if a.db == nil {
		return nil
	}
	return a.db.Close()
}

func (a *app) init() error {
	var errs error

	errs = errors.Join(errs, a.initPostgres())

	return errs
}

func (a *app) initPostgres() error {
	dbConfig := postgres.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	var err error
	a.db, err = postgres.NewPostgres(dbConfig)
	if err != nil {
		return err
	}

	return a.db.Init("../foundation/db_migrations")
}
