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
	ID          uint64
	UserID      string
	Name        string
	PrimaryType string
	MediaType   *string
	Duration    time.Duration
	Date        time.Time
	Meta        interface{}
	CreatedAt   time.Time
	DeletedAt   *time.Time
}

func NewActivity() *Activity {
	return &Activity{
		Meta: make(map[string]interface{}),
	}
}
