package user

import (
	"database/sql"

	"github.com/romanWienicke/go-app-test/business/data/db/dbuser"
)

type User struct {
	conn *sql.DB
}

func NewUser(db *sql.DB) *User {
	return &User{
		conn: db,
	}
}

func (u *User) GetUserByID(id int) (dbuser.User, error) {
	// Implement database retrieval logic here
	r, err := u.conn.Query("select id, name, email from users where id = $1", id)
	if err != nil {
		return dbuser.User{}, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	var user dbuser.User
	// Process the result set
	r.Next()
	err = r.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return dbuser.User{}, err
	}
	return user, nil
}

func (u *User) CreateUser(user dbuser.User) error {
	// Implement database insertion logic here
	_, err := u.conn.Exec("insert into users(name, email) values ($1, $2)", user.Name, user.Email)
	return err
}
