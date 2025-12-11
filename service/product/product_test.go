package product

import (
	"context"
	"testing"

	test "github.com/romanWienicke/go-app-test/foundation/testing"
)

func TestCreateProduct(t *testing.T) {
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

	productService := NewProductService(db)

	newProduct := Product{
		Name:        "Test Product",
		Price:       19.99,
		Description: "Test Product Description",
	}

	ctx := context.Background()
	id, err := productService.CreateProduct(ctx, newProduct)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	retrievedProduct, err := productService.GetProductByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve product: %v", err)
	}
	if retrievedProduct.Name != newProduct.Name || retrievedProduct.Price != newProduct.Price || retrievedProduct.Description != newProduct.Description {
		t.Fatalf("Retrieved product does not match created product")
	}

	updatedProduct := *retrievedProduct
	updatedProduct.Name = "Updated Product"
	updatedProduct.Price = 29.99
	updatedProduct.Description = "Updated Product Description"

	err = productService.UpdateProduct(ctx, updatedProduct)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	finalProduct, err := productService.GetProductByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated product: %v", err)
	}
	if finalProduct.Name != "Updated Product" || finalProduct.Price != 29.99 || finalProduct.Description != "Updated Product Description" {
		t.Fatalf("Updated product does not match expected values")
	}

	err = productService.DeleteProduct(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	_, err = productService.GetProductByID(ctx, id)
	if err == nil {
		t.Fatalf("Expected error when retrieving deleted product, got nil")
	}
}
