package commands

import (
	"context"

	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

var validTimezones = []string{
	"Local",
	"UTC",
	"GMT",
	"CST",
	"EET",
	"WET",
	"CET",
	"EST",
	"MST",

	"Cuba",
	"Egypt",
	"Eire",
	"Greenwich",
	"Iceland",
	"Iran",
	"Israel",
	"Jamaica",
	"Japan",
	"Libya",
	"Poland",
	"Portugal",
	"PRC",
	"Singapore",
	"Turkey",

	"Asia/Shanghai",
	"Asia/Chongqing",
	"Asia/Harbin",
	"Asia/Urumqi",
	"Asia/Hong_Kong",
	"Asia/Macao",
	"Asia/Taipei",
	"Asia/Tokyo",
	"Asia/Saigon",
	"Asia/Seoul",
	"Asia/Bangkok",
	"Asia/Dubai",
	"America/New_York",
	"America/Los_Angeles",
	"America/Chicago",
	"Europe/Moscow",
	"Europe/London",
	"Europe/Berlin",
	"Europe/Paris",
	"Europe/Rome",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Australia/Darwin",
}

func isValidTimezone(timezone string) bool {
	for _, tz := range validTimezones {
		if timezone == tz {
			return true
		}
	}
	return false
}

var ConfigCommandData = &discordgo.ApplicationCommand{
	Name:        "config",
	Description: "Configure your timezone and active guilds",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "timezone",
			Description: "Set your timezone",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	},
}

type ConfigCommand struct {
	r *users.UserRepository
}

func NewConfigCommand(r *users.UserRepository) *ConfigCommand {
	return &ConfigCommand{r: r}
}

func (c *ConfigCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	switch data.Options[0].Name {
	case "timezone":
		timezone := discordutil.GetRequiredStringOption(data.Options, "timezone")

		if !isValidTimezone(timezone) {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Invalid timezone!",
				},
			})
		}

		err := c.r.SetUserTimezone(context.Background(), discordutil.GetInteractionUser(i).ID, timezone)
		if err != nil {
			return err
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Timezone updated!",
			},
		})
	}
	return nil
}
