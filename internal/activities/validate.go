package activities

import (
	"errors"
	"slices"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrInvalidNameLength  = errors.New("name must be of length <100")
	ErrInvalidMediaType   = errors.New("invalid media type")
	ErrInvalidPrimaryType = errors.New("invalid primary type")
	ErrInvalidGuildID     = errors.New("guild id should be a valid discord snowflake")
	ErrInvalidUserID      = errors.New("user id should be a valid discord snowflake")
)

func isSnowflakeValid(str string) bool {
	if len(str) > 20 {
		return false
	}

	_, err := discordgo.SnowflakeTimestamp(str)
	return err == nil
}

func ValidateExternalActivity(a *Activity) error {
	validMediaType := []string{
		ActivityMediaTypeAnime,
		ActivityMediaTypeBook,
		ActivityMediaTypeManga,
		ActivityMediaTypeVideo,
		ActivityMediaTypeVisualNovel,
	}

	if utf8.RuneCountInString(a.Name) > 100 {
		return ErrInvalidNameLength
	}

	if a.MediaType != nil && !slices.Contains(validMediaType, *a.MediaType) {
		return ErrInvalidMediaType
	}

	if a.PrimaryType != ActivityImmersionTypeListening && a.PrimaryType != ActivityImmersionTypeReading {
		return ErrInvalidPrimaryType
	}

	if a.GuildID != nil && !isSnowflakeValid(*a.GuildID) {
		return ErrInvalidGuildID
	}

	if !isSnowflakeValid(a.UserID) {
		return ErrInvalidUserID
	}

	return nil
}
