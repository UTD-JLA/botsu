package activities

import (
	"time"
)

const (
	ActivityImmersionTypeReading   = "reading"
	ActivityImmersionTypeListening = "listening"
)

const (
	ActivityMediaTypeManga       = "manga"
	ActivityMediaTypeAnime       = "anime"
	ActivityMediaTypeVideo       = "video"
	ActivityMediaTypeBook        = "book"
	ActivityMediaTypeVisualNovel = "visual_novel"
)

type Activity struct {
	ID          uint64        `json:"id"`
	UserID      string        `json:"user_id"`
	GuildID     *string       `json:"guild_id"`
	Name        string        `json:"name"`
	PrimaryType string        `json:"primary_type"`
	MediaType   *string       `json:"media_type"`
	Duration    time.Duration `json:"duration"`
	Date        time.Time     `json:"date"`
	Meta        interface{}   `json:"meta"`
	CreatedAt   time.Time     `json:"created_at"`
	DeletedAt   *time.Time    `json:"deleted_at"`
}

func NewActivity() *Activity {
	return &Activity{
		Meta: make(map[string]interface{}),
	}
}
