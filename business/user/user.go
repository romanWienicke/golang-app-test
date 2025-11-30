package user

import (
	"context"

	"github.com/romanWienicke/go-app-test/business/user/data"
	"github.com/romanWienicke/go-app-test/foundation/postgres"
)

type User struct {
	db *postgres.Db
}

func NewUser(db *postgres.Db) *User {
	return &User{
		db: db,
	}
}

func (u *User) GetUserByID(ctx context.Context, id int) (*data.DbUser, error) {
	return postgres.QueryOne[data.DbUser](ctx, u.db.GetDB(), "select id, name, email from users where id=$1", id)
}

func (u *User) CreateUser(ctx context.Context, user data.DbUser) (int, error) {
	var id int
	err := u.db.GetDB().QueryRowContext(ctx,
		"insert into users (name, email) values ($1, $2) returning id;",
		user.Name, user.Email).Scan(&id)
	// You can use 'id' as needed, e.g., return it or log it
	return id, err
}
