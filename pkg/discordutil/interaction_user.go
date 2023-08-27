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

func IsSameInteractionUser(i1, i2 *discordgo.InteractionCreate) bool {
	return GetInteractionUser(i1).ID == GetInteractionUser(i2).ID
}
