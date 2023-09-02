package discordutil

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

const ResponseDeadline = 3 * time.Second
const InteractionTokenLifetime = 15 * time.Minute

func GetInteractionResponseDeadline(i *discordgo.Interaction) time.Time {
	// this shouldn't error, assuming snowflake is coming from Discord
	timestamp, _ := discordgo.SnowflakeTimestamp(i.ID)
	return timestamp.Add(ResponseDeadline)
}

func GetInteractionFollowupDeadline(i *discordgo.Interaction) time.Time {
	timestamp, _ := discordgo.SnowflakeTimestamp(i.ID)
	return timestamp.Add(InteractionTokenLifetime)
}
