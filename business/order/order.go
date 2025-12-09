package order

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Order struct {
	ID         uuid.UUID   `db:"id" json:"id" validate:"required,uuid4"`
	CustomerID uuid.UUID   `db:"customer_id" json:"customer_id" validate:"required,uuid4"`
	Status     string      `db:"status" json:"status" validate:"required"`
	Total      float64     `db:"total" json:"total" validate:"required,gt=0"`
	Items      []OrderItem `db:"items" json:"items" validate:"required,dive"`
}

func Validate(o Order) error {
	validator := validator.New()
	return validator.Struct(o)
}

type OrderItem struct {
	ID        uuid.UUID `db:"id" json:"id" validate:"required,uuid4"`
	OrderID   uuid.UUID `db:"order_id" json:"order_id" validate:"required,uuid4"`
	ProductID uuid.UUID `db:"product_id" json:"product_id" validate:"required,uuid4"`
	Quantity  float32   `db:"quantity" json:"quantity" validate:"required,gt=0"`
}

func ValidateItem(oi OrderItem) error {
	validator := validator.New()
	return validator.Struct(oi)
}
