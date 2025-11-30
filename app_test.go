package main

import (
	"database/sql"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/romanWienicke/go-app-test/business/data/db"
	"github.com/romanWienicke/go-app-test/docker"
	"github.com/romanWienicke/go-app-test/webtest"
)

func startServer(t *testing.T) string {
	port, err := webtest.GetRandomOpenPort(t)
	if err != nil || port == "" {
		port = "8080"
	}

	if err := os.Setenv("PORT", port); err != nil {
		t.Fatalf("Failed to set PORT environment variable: %v", err)
	}

	go main()
	// Allow some time for the server to start
	time.Sleep(1 * time.Second)

	return port
}

func startup(t *testing.T) *sql.DB {
	// Any necessary initialization before tests run
	composeFile := "docker-compose.yaml"
	dc, err := docker.ComposeUp(composeFile, "postgres")
	if err != nil {
		t.Fatalf("Failed to start Docker Compose: %v", err)
	}

	conn, err := db.ConnectPostgres("localhost", dc["postgres"].HostPorts["5432"], "root", "root", "testDb")
	if err != nil {
		t.Fatalf("Failed to connect to Postgres database: %v", err)
	}

	if err := db.InitDb(conn); err != nil {
		t.Fatalf("Failed to initialize Postgres database: %v", err)
	}

	return conn
}

func Test_Application(t *testing.T) {
	conn := startup(t)
	t.Cleanup(func() {
		t.Helper()

		if err := docker.ComposeDown("docker-compose.yaml"); err != nil {
			t.Errorf("Failed to stop Docker Compose: %v", err)
		}
	})

	_ = conn // Use conn in your tests

	port := startServer(t)
	tester := webtest.NewWebTest(port)

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
