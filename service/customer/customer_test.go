package customer

import (
	"context"
	"testing"

	test "github.com/romanWienicke/go-app-test/foundation/testing"
	"github.com/rs/zerolog"
)

func TestCustomerService(t *testing.T) {
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

	ctx := context.Background()
	log := zerolog.Nop()
	customerService := NewCustomerService(db, &log)

	// Test CreateCustomer
	newCustomer := Customer{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}
	id, err := customerService.CreateCustomer(ctx, newCustomer)
	if err != nil {
		t.Fatalf("Failed to create customer: %v", err)
	}

	// Test GetCustomerByID
	retrievedCustomer, err := customerService.GetCustomerByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve customer: %v", err)
	}
	if retrievedCustomer.Name != newCustomer.Name || retrievedCustomer.Email != newCustomer.Email {
		t.Fatalf("Retrieved customer does not match created customer")
	}

	// Test UpdateCustomer
	retrievedCustomer.Name = "Jane Doe"
	retrievedCustomer.Email = "jane.doe@example.com"
	err = customerService.UpdateCustomer(ctx, *retrievedCustomer)
	if err != nil {
		t.Fatalf("Failed to update customer: %v", err)
	}

	updatedCustomer, err := customerService.GetCustomerByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated customer: %v", err)
	}
	if updatedCustomer.Name != "Jane Doe" || updatedCustomer.Email != "jane.doe@example.com" {
		t.Fatalf("Updated customer does not match expected values")
	}

	// Test DeleteCustomer
	err = customerService.DeleteCustomer(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete customer: %v", err)
	}

	_, err = customerService.GetCustomerByID(ctx, id)
	if err == nil {
		t.Fatalf("Expected error when retrieving deleted customer, got none")
	}
}
