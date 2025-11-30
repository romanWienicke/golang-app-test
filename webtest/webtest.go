package webtest

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/romanWienicke/go-app-test/webtest/stopwatch"
)

type TestCase struct {
	Method       string
	Path         string
	Payload      any
	ExpectedCode int
	ExpectedBody any
}

type WebTest struct {
	port    string
	client  *http.Client
	baseURL string
}

type ResponseWithTime struct {
	Response *http.Response
	Elapsed  time.Duration
}

func NewWebTest(port string) *WebTest {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://localhost:%s/", port)

	return &WebTest{port: port, client: client, baseURL: url}
}

func (app *WebTest) RunTest(t *testing.T, tc TestCase) time.Duration {
	url := app.baseURL + strings.TrimPrefix(tc.Path, "/")
	resp := app.request(t, url, tc.Method, tc.Payload)
	app.expect(t, resp, tc.ExpectedCode, tc.ExpectedBody)
	return resp.Elapsed
}

func (app *WebTest) request(t *testing.T, url, method string, payload any) *ResponseWithTime {
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

func (app *WebTest) expect(t *testing.T, resp *ResponseWithTime, expectedStatus int, expectedBody any) {
	defer func() {
		if resp.Response.Body != nil {
			if err := resp.Response.Body.Close(); err != nil {
				t.Errorf("Failed to close response body: %v", err)
			}
		}
	}()

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

// GetRandomOpenPort returns a random open port as a string
func GetRandomOpenPort(t *testing.T) (string, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Fatalf("Failed to close listener: %v", err)
		}
	}()
	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}
