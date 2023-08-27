package discordutil

import "github.com/bwmarrin/discordgo"

// Because member is nil when not in a guild
// and I am going to cry if I get one more nil pointer error
func GetInteractionUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}
