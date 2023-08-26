package users

type User struct {
	ID           string
	ActiveGuilds []string
	Timezone     *string
}

func NewUser(id string) *User {
	return &User{
		ID:           id,
		ActiveGuilds: []string{},
	}
}
