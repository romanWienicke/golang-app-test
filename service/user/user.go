package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserService struct {
	db *postgres.Db
}

func NewUserService(db *postgres.Db) *UserService {
	return &UserService{
		db: db,
	}
}

func (u *UserService) RouteAdder() func(e *echo.Echo) {
	return func(e *echo.Echo) {
		e.POST("/user", func(c echo.Context) error {
			var newUser User
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

func (u *UserService) GetUserByID(ctx context.Context, id int) (*User, error) {
	return postgres.QueryOne[User](ctx, u.db.GetDB(), "select id, name, email from users where id=$1", id)
}

func (u *UserService) CreateUser(ctx context.Context, user User) (int, error) {
	var id int
	err := u.db.GetDB().QueryRowContext(ctx,
		"insert into users (name, email) values ($1, $2) returning id;",
		user.Name, user.Email).Scan(&id)
	// You can use 'id' as needed, e.g., return it or log it
	return id, err
}
