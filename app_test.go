package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/romanWienicke/go-app-test/stopwatch"
)

type TestCase struct {
	Method       string
	Path         string
	Payload      any
	ExpectedCode int
	ExpectedBody any
}

type AppTest struct {
	port    string
	client  *http.Client
	baseURL string
}

type ResponseWithTime struct {
	Response *http.Response
	Elapsed  time.Duration
}

func NewAppTest(port string) *AppTest {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://localhost:%s/", port)

	return &AppTest{port: port, client: client, baseURL: url}
}

func (app *AppTest) Test(t *testing.T, tc TestCase) time.Duration {
	url := app.baseURL + strings.TrimPrefix(tc.Path, "/")
	resp := app.request(t, url, tc.Method, tc.Payload)
	app.expect(t, resp, tc.ExpectedCode, tc.ExpectedBody)
	return resp.Elapsed
}

func (app *AppTest) request(t *testing.T, url, method string, payload any) *ResponseWithTime {
	client := app.client

	if payload != nil {
		var bodyReader io.Reader

		switch v := payload.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case map[string]interface{}:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}
			bodyReader = strings.NewReader(string(jsonBytes))
		default:
			t.Fatalf("Unsupported payload type: %T", payload)
		}

		contentType := "application/json"

		req, err := http.NewRequest(method, url, bodyReader)
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", contentType)

		sw := stopwatch.Stopwatch{}
		sw.Start()
		httpResp, err := client.Do(req)
		sw.Stop()
		if err != nil {
			t.Fatalf("Failed to perform request: %v", err)
		}

		return &ResponseWithTime{Response: httpResp, Elapsed: sw.Elapsed()}
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Errorf("Failed to create request: %v", err)
	}
	stopwatch := stopwatch.Stopwatch{}
	stopwatch.Start()
	resp, err := client.Do(req)
	stopwatch.Stop()
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}

	return &ResponseWithTime{Response: resp, Elapsed: stopwatch.Elapsed()}
}

func (app *AppTest) expect(t *testing.T, resp *ResponseWithTime, expectedStatus int, expectedBody any) {
	defer resp.Response.Body.Close()
	if resp.Response.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.Response.StatusCode)
	}

	if expectedBody != nil {
		bodyBytes, err := io.ReadAll(resp.Response.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		bodyString := string(bodyBytes)
		bodyString = strings.TrimSpace(bodyString)

		if bodyString != expectedBody {
			t.Errorf("Expected body %v, got %v", expectedBody, bodyString)
		}
	}
}

func startServer() string {
	port, err := getRandomOpenPort()
	if err != nil || port == "" {
		port = "8080"
	}
	os.Setenv("PORT", port)

	go main()
	// Allow some time for the server to start
	time.Sleep(1 * time.Second)

	return port
}

func Test_Application(t *testing.T) {
	port := startServer()
	app := NewAppTest(port)

	tests := map[string]TestCase{
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
			duration := app.Test(t, tc)
			t.Logf("Request %s took %v", name, duration)
		})
	}
}

// getRandomOpenPort returns a random open port as a string
func getRandomOpenPort() (string, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}
