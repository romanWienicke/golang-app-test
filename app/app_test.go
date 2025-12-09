package app

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	test "github.com/romanWienicke/go-app-test/foundation/testing"
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
	test.SetEnv(t, "../.env")
	dc := test.DockerComposeUp(t, "../docker-compose.yaml")

	if err := os.Setenv("DB_HOST", "localhost"); err != nil {
		t.Fatalf("Failed to set DB_HOST environment variable: %v", err)
	}
	test.SetupDatabaseEnv(t, dc["postgres"])
}

func Test_Application(t *testing.T) {
	startup(t)

	t.Cleanup(func() {
		t.Helper()
		test.DockerComposeDown(t, "../docker-compose.yaml")
	})

	tester := webtest.NewWebTest(startServer(t))
	tests := []struct {
		name string
		tc   webtest.TestCase
	}{
		{"GET /", webtest.TestCase{
			Method:       http.MethodGet,
			Path:         "/",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "Hello, world!",
		}},
		{"POST / with JSON", webtest.TestCase{
			Method:       http.MethodPost,
			Path:         "/",
			Payload:      map[string]interface{}{"key": "value"},
			ExpectedCode: http.StatusCreated,
			ExpectedBody: `{"message":"JSON received"}`,
		}},
		{"PUT / with JSON", webtest.TestCase{
			Method:       http.MethodPut,
			Path:         "/",
			Payload:      map[string]interface{}{"key": "value"},
			ExpectedCode: http.StatusAccepted,
			ExpectedBody: `{"message":"JSON received"}`,
		}},
		{"DELETE /any/:id", webtest.TestCase{
			Method:       http.MethodDelete,
			Path:         "/any/123",
			ExpectedCode: http.StatusOK,
			ExpectedBody: `{"id":"123","message":"Resource deleted"}`,
		}},
		{"GET /ping", webtest.TestCase{
			Method:       http.MethodGet,
			Path:         "/ping",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "pong",
		}},
		{"GET /notfound", webtest.TestCase{
			Method:       http.MethodGet,
			Path:         "/notfound",
			ExpectedCode: http.StatusNotFound,
		}},
		{"DELETE / without id", webtest.TestCase{
			Method:       http.MethodDelete,
			Path:         "/any/",
			ExpectedCode: http.StatusNotFound,
		}},
		{"POST / with invalid JSON", webtest.TestCase{
			Method:       http.MethodPost,
			Path:         "/",
			Payload:      "invalid json",
			ExpectedCode: http.StatusBadRequest,
			ExpectedBody: `{"error":"Invalid request body"}`,
		}},
		{"POST /user with valid data", webtest.TestCase{
			Method:              http.MethodPost,
			Path:                "/user",
			Payload:             map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			ExpectedCode:        http.StatusCreated,
			ExpectedBodyPattern: "{\"id\":(?P<id>\\d+),\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
		}},
		{"GET /user/:id", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/user/:id",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\\d+,\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			duration := tester.RunTest(t, tc.tc)
			t.Logf("Request %s took %v", tc.name, duration)
		})
	}
}
