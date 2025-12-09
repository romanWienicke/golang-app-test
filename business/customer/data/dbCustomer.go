package data

import "github.com/google/uuid"

type DbCustomer struct {
	Id    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
