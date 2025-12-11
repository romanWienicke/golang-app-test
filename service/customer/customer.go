package customer

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
	"github.com/rs/zerolog"
)

type Customer struct {
	ID    uuid.UUID `db:"id" json:"id" validate:"omitempty,uuid4"`
	Name  string    `db:"name" json:"name" validate:"required,min=2,max=100"`
	Email string    `db:"email" json:"email" validate:"required,email"`
}

func Validate(c Customer) error {
	validator := validator.New()
	return validator.Struct(c)
}

type CustomerService struct {
	db  *postgres.Db
	log *zerolog.Logger
}

func NewCustomerService(db *postgres.Db, log *zerolog.Logger) *CustomerService {
	return &CustomerService{
		db:  db,
		log: log,
	}
}

func (cs *CustomerService) RouteAdder() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		e.POST("/customer", func(c echo.Context) error {
			var newCustomer Customer
			if err := c.Bind(&newCustomer); err != nil {
				cs.log.Error().Err(err).Msg("Failed to bind customer")
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}

			id, err := cs.CreateCustomer(c.Request().Context(), newCustomer)
			if err != nil {
				cs.log.Error().Err(err).Msg("Failed to create customer")
				return c.JSON(500, map[string]string{"error": "Failed to create customer"})
			}

			newCustomer.ID = id
			return c.JSON(201, newCustomer)
		})
		e.GET("/customer/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				cs.log.Error().Err(err).Msg("Invalid customer ID")
				return c.JSON(400, map[string]string{"error": "Invalid customer ID"})
			}

			customer, err := cs.GetCustomerByID(c.Request().Context(), id)
			if err != nil {
				if err == postgres.ErrNoRows {
					cs.log.Error().Err(err).Msg("Customer not found")
					return c.JSON(404, map[string]string{"error": "Customer not found"})
				}
				cs.log.Error().Err(err).Msg("Failed to retrieve customer")
				return c.JSON(500, map[string]string{"error": "Failed to retrieve customer"})
			}
			return c.JSON(200, customer)
		})
		e.PUT("/customer/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				cs.log.Error().Err(err).Msg("Invalid customer ID")
				return c.JSON(400, map[string]string{"error": "Invalid customer ID"})
			}

			var updatedCustomer Customer
			if err := c.Bind(&updatedCustomer); err != nil {
				cs.log.Error().Err(err).Msg("Invalid request body")
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}
			updatedCustomer.ID = id

			if err := cs.UpdateCustomer(c.Request().Context(), updatedCustomer); err != nil {
				cs.log.Error().Err(err).Msg("Failed to update customer")
				return c.JSON(500, map[string]string{"error": "Failed to update customer"})
			}
			return c.JSON(200, updatedCustomer)
		})
		e.DELETE("/customer/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				cs.log.Error().Err(err).Msg("Invalid customer ID")
				return c.JSON(400, map[string]string{"error": "Invalid customer ID"})
			}

			if err := cs.DeleteCustomer(c.Request().Context(), id); err != nil {
				cs.log.Error().Err(err).Msg("Failed to delete customer")
				return c.JSON(500, map[string]string{"error": "Failed to delete customer"})
			}
			return c.NoContent(204)
		})
	}
}

func (cs *CustomerService) CreateCustomer(ctx context.Context, customer Customer) (uuid.UUID, error) {
	if err := Validate(customer); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	_, err := cs.db.GetDB().ExecContext(ctx,
		"insert into customers (id, name, email) values ($1, $2, $3);",
		id, customer.Name, customer.Email)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (cs *CustomerService) GetCustomerByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return postgres.QueryOne[Customer](ctx, cs.db.GetDB(), "select id, name, email from customers where id=$1", id)
}

func (cs *CustomerService) UpdateCustomer(ctx context.Context, customer Customer) error {
	if err := Validate(customer); err != nil {
		return err
	}

	_, err := cs.db.GetDB().ExecContext(ctx,
		"update customers set name=$1, email=$2 where id=$3;",
		customer.Name, customer.Email, customer.ID)
	return err
}

func (cs *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	_, err := cs.db.GetDB().ExecContext(ctx,
		"delete from customers where id=$1;",
		id)
	return err
}
