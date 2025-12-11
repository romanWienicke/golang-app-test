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
		{"POST /user with valid data", webtest.TestCase{
			Method:              http.MethodPost,
			Path:                "/user",
			Payload:             map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
			ExpectedCode:        http.StatusCreated,
			ExpectedBodyPattern: "{\"id\":(?P<userId>\\d+),\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
		}},
		{"GET /user/:userId", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/user/:userId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\\d+,\"name\":\"Alice\",\"email\":\"alice@example.com\"}",
		}},
		{"POST /customer with valid data", webtest.TestCase{
			Method:              http.MethodPost,
			Path:                "/customer",
			Payload:             map[string]interface{}{"name": "Bob", "email": "bob@example.com"},
			ExpectedCode:        http.StatusCreated,
			ExpectedBodyPattern: "{\"id\":\"(?P<customerId>[0-9a-fA-F-]{36})\",\"name\":\"Bob\",\"email\":\"bob@example.com\"}",
		}},
		{"GET /customer/:customerId", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/customer/:customerId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Bob\",\"email\":\"bob@example.com\"}",
		}},
		{"PUT /customer/:customerId", webtest.TestCase{
			Method:              http.MethodPut,
			Path:                "/customer/:customerId",
			Payload:             map[string]interface{}{"name": "Robert", "email": "robert@example.com"},
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Robert\",\"email\":\"robert@example.com\"}",
		}},
		{"GET /customer/:customerId after update", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/customer/:customerId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Robert\",\"email\":\"robert@example.com\"}",
		}},

		{"POST /product with valid data", webtest.TestCase{
			Method:              http.MethodPost,
			Path:                "/product",
			Payload:             map[string]interface{}{"name": "Widget", "price": 19.99},
			ExpectedCode:        http.StatusCreated,
			ExpectedBodyPattern: "{\"id\":\"(?P<productId>[0-9a-fA-F-]{36})\",\"name\":\"Widget\",\"description\":\"\",\"price\":19.99}",
		}},
		{"GET /product/:productId", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/product/:productId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Widget\",\"description\":\"\",\"price\":19.99}",
		}},
		{"PUT /product/:productId", webtest.TestCase{
			Method:              http.MethodPut,
			Path:                "/product/:productId",
			Payload:             map[string]interface{}{"name": "Super Widget", "description": "An improved widget", "price": 29.99},
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Super Widget\",\"description\":\"An improved widget\",\"price\":29.99}",
		}},
		{"GET /product/:productId after update", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/product/:productId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"name\":\"Super Widget\",\"description\":\"An improved widget\",\"price\":29.99}",
		}},

		{"POST /order with valid data", webtest.TestCase{
			Method: http.MethodPost,
			Path:   "/order",
			Payload: map[string]interface{}{"customer_id": ":customerId", "status": "pending", "total": 39.98, "items": []map[string]interface{}{
				{"product_id": ":productId", "quantity": 2},
			}},
			ExpectedCode:        http.StatusCreated,
			ExpectedBodyPattern: "{\"id\":\"(?P<orderId>[0-9a-fA-F-]{36})\",\"customer_id\":\":customerId\",\"status\":\"pending\",\"total\":39.98,\"items\":\\[\\{\"id\":\"[0-9a-fA-F-]{36}\",\"order_id\":\"[0-9a-fA-F-]{36}\",\"product_id\":\":productId\",\"quantity\":2\\}\\]\\}",
		}},
		{"GET /order/:orderId", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/order/:orderId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\":orderId\",\"customer_id\":\":customerId\",\"status\":\"pending\",\"total\":39.98,\"items\":\\[\\{\"id\":\"[0-9a-fA-F-]{36}\",\"order_id\":\"[0-9a-fA-F-]{36}\",\"product_id\":\":productId\",\"quantity\":2}\\]\\}",
		}},
		{"PUT /order/:orderId", webtest.TestCase{
			Method: http.MethodPut,
			Path:   "/order/:orderId",
			Payload: map[string]interface{}{"customer_id": ":customerId", "status": "pending", "total": 11.98, "items": []map[string]interface{}{
				{"product_id": ":productId", "quantity": 3},
			}},
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\"[0-9a-fA-F-]{36}\",\"customer_id\":\":customerId\",\"status\":\"pending\",\"total\":11.98,\"items\":\\[\\{\"id\":\"[0-9a-fA-F-]{36}\",\"order_id\":\"[0-9a-fA-F-]{36}\",\"product_id\":\":productId\",\"quantity\":3\\}\\]\\}",
		}},
		{"GET /order/:orderId after update", webtest.TestCase{
			Method:              http.MethodGet,
			Path:                "/order/:orderId",
			ExpectedCode:        http.StatusOK,
			ExpectedBodyPattern: "{\"id\":\":orderId\",\"customer_id\":\":customerId\",\"status\":\"pending\",\"total\":11.98,\"items\":\\[\\{\"id\":\"[0-9a-fA-F-]{36}\",\"order_id\":\"[0-9a-fA-F-]{36}\",\"product_id\":\":productId\",\"quantity\":3\\}\\]\\}",
		}},
		{"DELETE /order/:orderId", webtest.TestCase{
			Method:       http.MethodDelete,
			Path:         "/order/:orderId",
			ExpectedCode: http.StatusNoContent,
		}},
		{"DELETE /product/:productId", webtest.TestCase{
			Method:       http.MethodDelete,
			Path:         "/product/:productId",
			ExpectedCode: http.StatusNoContent,
		}},
		{"DELETE /customer/:customerId", webtest.TestCase{
			Method:       http.MethodDelete,
			Path:         "/customer/:customerId",
			ExpectedCode: http.StatusNoContent,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			duration := tester.RunTest(t, tc.tc)
			t.Logf("Request %s took %v", tc.name, duration)
		})
	}
}
