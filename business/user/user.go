package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/romanWienicke/go-app-test/business/user/data"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

type UserBuis struct {
	db *postgres.Db
}

func NewUser(db *postgres.Db) *UserBuis {
	return &UserBuis{
		db: db,
	}
}

func (u *UserBuis) RouteAdder() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		e.POST("/user", func(c echo.Context) error {
			var newUser data.DbUser
			if err := c.Bind(&newUser); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			}

			id, err := u.CreateUser(c.Request().Context(), newUser)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
			}

			newUser.Id = id
			return c.JSON(http.StatusCreated, newUser)
		})

		e.GET("/user/:id", func(c echo.Context) error {
			idParam := c.Param("id")
			id, err := strconv.Atoi(idParam)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
			}

			user, err := u.GetUserByID(c.Request().Context(), id)
			if err != nil {
				if errors.Is(err, postgres.ErrNoRows) {
					return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
				}
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user"})
			}

			return c.JSON(http.StatusOK, user)
		})
	}
}

func (u *UserBuis) GetUserByID(ctx context.Context, id int) (*data.DbUser, error) {
	return postgres.QueryOne[data.DbUser](ctx, u.db.GetDB(), "select id, name, email from users where id=$1", id)
}

func (u *UserBuis) CreateUser(ctx context.Context, user data.DbUser) (int, error) {
	var id int
	err := u.db.GetDB().QueryRowContext(ctx,
		"insert into users (name, email) values ($1, $2) returning id;",
		user.Name, user.Email).Scan(&id)
	// You can use 'id' as needed, e.g., return it or log it
	return id, err
}
