package app

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/romanWienicke/go-app-test/docker"
	"github.com/romanWienicke/go-app-test/webtest"
)

func startServer(t *testing.T) string {
	port, err := webtest.GetRandomOpenPort(t)
	if err != nil || port == "" {
		port = "8080"
	}

	if err := os.Setenv("HTTP_PORT", port); err != nil {
		t.Fatalf("Failed to set PORT environment variable: %v", err)
	}

	go func() {
		a, err := NewApp()
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize app: %v", err))
		}

		a.Run()
	}()

	// Allow some time for the server to start
	time.Sleep(1 * time.Second)

	return port
}

func startup(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("No .env file found (continuing)")
	}

	// Any necessary initialization before tests run
	composeFile := "../docker-compose.yaml"
	dc, err := docker.ComposeUp(composeFile, "postgres")
	if err != nil {
		t.Fatalf("Failed to start Docker Compose: %v", err)
	}

	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", dc["postgres"].HostPorts["5432"])
}

func Test_Application(t *testing.T) {
	startup(t)

	t.Cleanup(func() {
		t.Helper()

		if err := docker.ComposeDown("../docker-compose.yaml"); err != nil {
			t.Errorf("Failed to stop Docker Compose: %v", err)
		}
	})

	tester := webtest.NewWebTest(startServer(t))
	tests := map[string]webtest.TestCase{
		"GET /": {
			Method:       http.MethodGet,
			Path:         "/",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "Hello, world!",
		},
		"POST / with JSON": {
			Method:       http.MethodPost,
			Path:         "/",
			Payload:      map[string]interface{}{"key": "value"},
			ExpectedCode: http.StatusCreated,
			ExpectedBody: `{"message":"JSON received"}`,
		},
		"PUT / with JSON": {
			Method:       http.MethodPut,
			Path:         "/",
			Payload:      map[string]interface{}{"key": "value"},
			ExpectedCode: http.StatusAccepted,
			ExpectedBody: `{"message":"JSON received"}`,
		},
		"DELETE /any/:id": {
			Method:       http.MethodDelete,
			Path:         "/any/123",
			ExpectedCode: http.StatusOK,
			ExpectedBody: `{"id":"123","message":"Resource deleted"}`,
		},
		"GET /ping": {
			Method:       http.MethodGet,
			Path:         "/ping",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "pong",
		},
		"GET /notfound": {
			Method:       http.MethodGet,
			Path:         "/notfound",
			ExpectedCode: http.StatusNotFound,
		},
		"DELETE / without id": {
			Method:       http.MethodDelete,
			Path:         "/any/",
			ExpectedCode: http.StatusNotFound,
		},
		"POST / with invalid JSON": {
			Method:       http.MethodPost,
			Path:         "/",
			Payload:      "invalid json",
			ExpectedCode: http.StatusBadRequest,
			ExpectedBody: `{"error":"Invalid request body"}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			duration := tester.RunTest(t, tc)
			t.Logf("Request %s took %v", name, duration)
		})
	}
}
