package users

type User struct {
	ID       string
	Timezone *string
}

func NewUser(id string) *User {
	return &User{
		ID:       id,
		Timezone: nil,
	}
}
