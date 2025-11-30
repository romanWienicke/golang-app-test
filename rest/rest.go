package rest

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type Payload struct {
	Key string `json:"key" validate:"required"`
}

func NewServer(port string) error {
	e := echo.New()
	validator := validator.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, world!")
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.POST("/", func(c echo.Context) error {
		var data Payload

		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if err := validator.Struct(data); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		return c.JSON(http.StatusCreated, map[string]string{"message": "JSON received"})
	})

	e.PUT("/", func(c echo.Context) error {
		var body Payload
		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		if err := validator.Struct(body); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}

		return c.JSON(http.StatusAccepted, map[string]string{"message": "JSON received"})
	})

	e.DELETE("/any/:id", func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing id parameter"})
		}
		// Here you would handle deletion logic using the id
		return c.JSON(http.StatusOK, map[string]string{"message": "Resource deleted", "id": id})
	})

	return e.Start(":" + port)
}
