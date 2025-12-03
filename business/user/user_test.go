package user

import (
	"context"
	"testing"

	"github.com/romanWienicke/go-app-test/business/user/data"
	test "github.com/romanWienicke/go-app-test/foundation/testing"
)

func TestCreateUser(t *testing.T) {
	test.SetEnv(t, "../../.env")
	test.DockerComposeUp(t, "../../docker-compose.yaml", "postgres")
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

	userService := NewUser(db)

	newUser := data.DbUser{
		Name:  "Test User",
		Email: "testuser@example.com",
	}

	ctx := context.Background()

	id, err := userService.CreateUser(ctx, newUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	retrievedUser, err := userService.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	if retrievedUser.Name != newUser.Name || retrievedUser.Email != newUser.Email {
		t.Fatalf("Retrieved user does not match created user")
	}
}
