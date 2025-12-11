package order

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

type Order struct {
	ID         uuid.UUID   `db:"id" json:"id" validate:"omitempty,uuid4"`
	CustomerID uuid.UUID   `db:"customer_id" json:"customer_id" validate:"required,uuid4"`
	Status     string      `db:"status" json:"status" validate:"required"`
	Total      float64     `db:"total" json:"total" validate:"required,gt=0"`
	Items      []OrderItem `db:"items" json:"items" validate:"required"`
}

func Validate(o Order) error {
	validator := validator.New()
	return validator.Struct(o)
}

type OrderItem struct {
	ID        uuid.UUID `db:"id" json:"id" validate:"omitempty,uuid4"`
	OrderID   uuid.UUID `db:"order_id" json:"order_id" validate:"required,uuid4"`
	ProductID uuid.UUID `db:"product_id" json:"product_id" validate:"required,uuid4"`
	Quantity  float32   `db:"quantity" json:"quantity" validate:"required,gt=0"`
}

func ValidateItem(oi OrderItem) error {
	validator := validator.New()
	return validator.Struct(oi)
}

type OrderService struct {
	db *postgres.Db
}

func NewOrderService(db *postgres.Db) *OrderService {
	return &OrderService{
		db: db,
	}
}

func (o *OrderService) RouteAdder() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		e.POST("/order", func(c echo.Context) error {
			var newOrder Order
			if err := c.Bind(&newOrder); err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}

			id, err := o.CreateOrder(c.Request().Context(), newOrder)
			if err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to create order"})
			}

			newOrder.ID = id
			return c.JSON(201, newOrder)
		})

		e.GET("/order/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid order ID"})
			}

			order, err := o.GetOrderByID(c.Request().Context(), id)
			if err != nil {
				if err == postgres.ErrNoRows {
					return c.JSON(404, map[string]string{"error": "Order not found"})
				}
				return c.JSON(500, map[string]string{"error": "Failed to retrieve order"})
			}
			return c.JSON(200, order)
		})

		e.PUT("/order/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid order ID"})
			}

			var updatedOrder Order
			if err := c.Bind(&updatedOrder); err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}
			updatedOrder.ID = id

			if err := o.UpdateOrder(c.Request().Context(), updatedOrder); err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to update order"})
			}
			return c.JSON(200, updatedOrder)
		})
		e.DELETE("/order/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid order ID"})
			}

			if err := o.DeleteOrder(c.Request().Context(), id); err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to delete order"})
			}
			return c.NoContent(204)
		})
	}
}

func (o *OrderService) CreateOrder(ctx context.Context, order Order) (uuid.UUID, error) {
	if err := Validate(order); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	_, err := o.db.GetDB().ExecContext(ctx,
		"insert into orders (id, customer_id, status, total) values ($1, $2, $3, $4);",
		id, order.CustomerID, order.Status, order.Total)
	if err != nil {
		return uuid.Nil, err
	}

	for _, item := range order.Items {
		item.OrderID = id
		if err := ValidateItem(item); err != nil {
			return uuid.Nil, err
		}
		itemID := uuid.New()
		_, err := o.db.GetDB().ExecContext(ctx,
			"insert into order_items (id, order_id, product_id, quantity) values ($1, $2, $3, $4);",
			itemID, id, item.ProductID, item.Quantity)
		if err != nil {
			return uuid.Nil, err
		}
	}

	return id, nil
}

func (o *OrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	order, err := postgres.QueryOne[Order](ctx, o.db.GetDB(), "select id, customer_id, status, total from orders where id=$1", id)
	if err != nil {
		return nil, err
	}

	items, err := postgres.QueryList[OrderItem](ctx, o.db.GetDB(), "select id, order_id, product_id, quantity from order_items where order_id=$1", id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

func (o *OrderService) UpdateOrder(ctx context.Context, order Order) error {
	if err := Validate(order); err != nil {
		return err
	}

	_, err := o.db.GetDB().ExecContext(ctx,
		"update orders set customer_id=$1, status=$2, total=$3 where id=$4;",
		order.CustomerID, order.Status, order.Total, order.ID)

	for _, item := range order.Items {
		item.OrderID = order.ID
		if err := ValidateItem(item); err != nil {
			return err
		}

		_, err := o.db.GetDB().ExecContext(ctx,
			"update order_items set quantity=$1 where order_id=$2 and product_id=$3;",
			item.Quantity, order.ID, item.ProductID)
		if err != nil {
			return err
		}
	}

	return err
}

func (o *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	_, err := o.db.GetDB().ExecContext(ctx,
		"delete from orders where id=$1;",
		id)
	return err
}
