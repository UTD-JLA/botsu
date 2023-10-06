package users

type User struct {
	ID                      string
	Timezone                *string
	VisualNovelReadingSpeed float32
	BookReadingSpeed        float32
	MangaReadingSpeed       float32
	DailyGoal               int
}

func NewUser(id string) *User {
	return &User{
		ID:                      id,
		Timezone:                nil,
		VisualNovelReadingSpeed: 0,
		BookReadingSpeed:        0,
		MangaReadingSpeed:       0,
		DailyGoal:               0,
	}
}
