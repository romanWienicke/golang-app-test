package user

import (
	"context"
	"os"
	"testing"

	"github.com/romanWienicke/go-app-test/business/user/data"
	"github.com/romanWienicke/go-app-test/docker"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

func TestCreateUser(t *testing.T) {
	dc, err := docker.ComposeUp("../../docker-compose.yaml", "postgres")
	if err != nil {
		t.Fatalf("Failed to start Docker Compose: %v", err)
	}
	defer docker.ComposeDown("../../docker-compose.yaml")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", dc["postgres"].HostPorts["5432"])

	db, err := initPostgres()
	if err != nil {
		t.Fatalf("Failed to initialize Postgres: %v", err)
	}
	defer db.Close()

	userService := NewUser(db)

	newUser := data.DbUser{
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	ctx := context.Background()

	id, err := userService.CreateUser(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrievedUser, err := userService.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	if retrievedUser.Name != newUser.Name || retrievedUser.Email != newUser.Email {
		t.Fatalf("Retrieved user does not match created user")
	}
}

func initPostgres() (*postgres.Db, error) {
	dbConfig := postgres.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	db, err := postgres.NewPostgres(dbConfig)
	if err != nil {
		return nil, err
	}

	if err := db.Init("../foundation/db_migrations"); err != nil {
		return nil, err
	}

	return db, nil
}
