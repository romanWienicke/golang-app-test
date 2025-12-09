package customer

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Customer struct {
	ID    uuid.UUID `db:"id" json:"id" validate:"required,uuid4"`
	Name  string    `db:"name" json:"name" validate:"required, min=2,max=100"`
	Email string    `db:"email" json:"email" validate:"required,email"`
}

func Validate(c Customer) error {
	validator := validator.New()
	return validator.Struct(c)
}
