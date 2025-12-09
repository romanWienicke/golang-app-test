package product

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `db:"id" json:"id" validate:"required,uuid4"`
	Name        string    `db:"name" json:"name" validate:"required, min=2,max=100"`
	Description string    `db:"description" json:"description" validate:"max=2000"`
	Price       float64   `db:"price" json:"price" validate:"required,gt=0"`
}

func Validate(p Product) error {
	validator := validator.New()
	return validator.Struct(p)
}
