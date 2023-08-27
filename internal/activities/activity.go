package activities

import (
	"database/sql/driver"
	"encoding/json"
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

const (
	ActivityMetaUnitNameBookPage  = "book_page"
	ActivityMetaUnitNameMangaPage = "manga_page"
	ActivityMetaUnitNameCharacter = "character"
	ActivityMetaUnitNameEpisode   = "episode"
)

type ActivityMeta map[string]interface{}

func (m ActivityMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *ActivityMeta) Scan(src interface{}) error {
	return json.Unmarshal([]byte(src.(string)), m)
}

type Activity struct {
	ID          uint64
	UserID      string
	Name        string
	PrimaryType string
	MediaType   *string
	Duration    time.Duration
	Date        time.Time
	Meta        ActivityMeta
}

func NewActivity() *Activity {
	return &Activity{
		Meta: make(ActivityMeta),
	}
}
