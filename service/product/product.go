package product

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

type Product struct {
	ID          uuid.UUID `db:"id" json:"id" validate:"omitempty,uuid4"`
	Name        string    `db:"name" json:"name" validate:"required,min=2,max=100"`
	Description string    `db:"description" json:"description" validate:"max=2000"`
	Price       float64   `db:"price" json:"price" validate:"required,gt=0"`
}

func Validate(p Product) error {
	validator := validator.New()
	return validator.Struct(p)
}

type ProductService struct {
	db *postgres.Db
}

func NewProductService(db *postgres.Db) *ProductService {
	return &ProductService{
		db: db,
	}
}

func (p *ProductService) RouteAdder() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		e.POST("/product", func(c echo.Context) error {
			var newProduct Product
			if err := c.Bind(&newProduct); err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}

			id, err := p.CreateProduct(c.Request().Context(), newProduct)
			if err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to create product"})
			}

			newProduct.ID = id
			return c.JSON(201, newProduct)
		})

		e.GET("/product/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid product ID"})
			}

			product, err := p.GetProductByID(c.Request().Context(), id)
			if err != nil {
				if err == postgres.ErrNoRows {
					return c.JSON(404, map[string]string{"error": "Product not found"})
				}
				return c.JSON(500, map[string]string{"error": "Failed to retrieve product"})
			}
			return c.JSON(200, product)
		})

		e.PUT("/product/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid product ID"})
			}

			var updatedProduct Product
			if err := c.Bind(&updatedProduct); err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid request body"})
			}
			updatedProduct.ID = id

			if err := p.UpdateProduct(c.Request().Context(), updatedProduct); err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to update product"})
			}
			return c.NoContent(204)
		})

		e.DELETE("/product/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := uuid.Parse(idParam)
			if err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid product ID"})
			}

			if err := p.DeleteProduct(c.Request().Context(), id); err != nil {
				return c.JSON(500, map[string]string{"error": "Failed to delete product"})
			}
			return c.NoContent(204)
		})
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, product Product) (uuid.UUID, error) {
	if err := Validate(product); err != nil {
		return uuid.Nil, err
	}

	id := uuid.New()
	_, err := p.db.GetDB().ExecContext(ctx,
		"insert into products (id, name, description, price) values ($1, $2, $3, $4);",
		id, product.Name, product.Description, product.Price)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (p *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	return postgres.QueryOne[Product](ctx, p.db.GetDB(), "select id, name, description, price from products where id=$1", id)
}

func (p *ProductService) UpdateProduct(ctx context.Context, product Product) error {
	if err := Validate(product); err != nil {
		return err
	}

	_, err := p.db.GetDB().ExecContext(ctx,
		"update products set name=$1, description=$2, price=$3 where id=$4;",
		product.Name, product.Description, product.Price, product.ID)
	return err
}

func (p *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.GetDB().ExecContext(ctx,
		"delete from products where id=$1;",
		id)
	return err
}
