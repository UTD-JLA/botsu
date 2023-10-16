package guilds

type Guild struct {
	ID       string
	Timezone *string
}

func NewGuild(id string) *Guild {
	return &Guild{
		ID:       id,
		Timezone: nil,
	}
}
