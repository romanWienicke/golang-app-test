package dbuser

type User struct {
	ID    int
	Name  string
	Email string
}

func NewUser(id int, name, email string) *User {
	return &User{
		ID:    id,
		Name:  name,
		Email: email,
	}
}
