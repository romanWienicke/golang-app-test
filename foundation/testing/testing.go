package testing

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/romanWienicke/go-app-test/docker"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

func SetEnv(t *testing.T, envFile string) {
	if err := godotenv.Load(envFile); err != nil {
		t.Fatalf("Failed to load .env file: %v", err)
	}
}

func DockerComposeUp(t *testing.T, composeFile string, services ...string) map[string]docker.Container {
	wd, _ := os.Getwd()
	t.Logf("Current working directory: %s", wd)

	dc, err := docker.ComposeUp(t, composeFile, services...)
	if err != nil {
		t.Fatalf("Failed to start Docker Compose: %v", err)
	}

	return dc
}

func SetupDatabaseEnv(t *testing.T, container docker.Container) {
	if err := os.Setenv("DB_HOST", "localhost"); err != nil {
		t.Fatalf("Failed to set DB_HOST environment variable: %v", err)
	}
	if err := os.Setenv("DB_PORT", container.HostPorts["5432"]); err != nil {
		t.Fatalf("Failed to set DB_PORT environment variable: %v", err)
	}
}

func DockerComposeDown(t *testing.T, composeFile string) {
	if err := docker.ComposeDown(t, composeFile); err != nil {
		t.Fatalf("Failed to stop Docker Compose: %v", err)
	}
}

func InitPostgres(t *testing.T, migrationFile string) *postgres.Db {
	dbConfig := postgres.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
	db, err := postgres.NewPostgres(dbConfig)
	if err != nil {
		t.Fatalf("Failed to initialize Postgres: %v", err)
	}

	if err := db.Init(migrationFile); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}
