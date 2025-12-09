package order

import (
	"context"
	"testing"

	test "github.com/romanWienicke/go-app-test/foundation/testing"
	"github.com/romanWienicke/go-app-test/service/customer"
	"github.com/romanWienicke/go-app-test/service/product"
)

func TestCreateOrder(t *testing.T) {
	test.SetEnv(t, "../../.env")
	dc := test.DockerComposeUp(t, "../../docker-compose.yaml", "postgres")
	test.SetupDatabaseEnv(t, dc["postgres"])
	db := test.InitPostgres(t, "../../migrations")
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close db: %v", err)
		}
	}()

	t.Cleanup(func() {
		t.Helper()
		test.DockerComposeDown(t, "../../docker-compose.yaml")
	})

	customerService := customer.NewCustomerService(db)
	// Create a customer to associate with the product if needed
	customerID, err := customerService.CreateCustomer(context.Background(), customer.Customer{
		Name:  "Test Customer",
		Email: "testcustomer@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	productService := product.NewProductService(db)
	// Create a product to associate with the order
	productID, err := productService.CreateProduct(context.Background(), product.Product{
		Name:        "Test Product",
		Price:       24.99,
		Description: "Test Product Description",
	})
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	orderService := NewOrderService(db)
	newOrder := Order{
		CustomerID: customerID,
		Status:     "pending",
		Total:      49.99,
		Items: []OrderItem{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	ctx := context.Background()
	id, err := orderService.CreateOrder(ctx, newOrder)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	retrievedOrder, err := orderService.GetOrderByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve order: %v", err)
	}
	if retrievedOrder.CustomerID != newOrder.CustomerID || len(retrievedOrder.Items) != len(newOrder.Items) {
		t.Fatalf("Retrieved order does not match created order")
	}

	updatedOrder := *retrievedOrder
	updatedOrder.Status = "shipped"
	updatedOrder.Total = 99.99

	err = orderService.UpdateOrder(ctx, updatedOrder)
	if err != nil {
		t.Fatalf("Failed to update order: %v", err)
	}

	finalOrder, err := orderService.GetOrderByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated order: %v", err)
	}
	if finalOrder.Status != updatedOrder.Status || finalOrder.Total != updatedOrder.Total {
		t.Fatalf("Retrieved order does not match updated order")
	}

	err = orderService.DeleteOrder(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete order: %v", err)
	}
}
