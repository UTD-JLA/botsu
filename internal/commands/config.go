package commands

import (
	"strings"
	"unicode"

	"github.com/UTD-JLA/botsu/internal/bot"
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
			Name:         "timezone",
			Description:  "Set your timezone",
			Type:         discordgo.ApplicationCommandOptionString,
			Required:     false,
			Autocomplete: true,
		},
	},
}

type ConfigCommand struct {
	r *users.UserRepository
}

func NewConfigCommand(r *users.UserRepository) *ConfigCommand {
	return &ConfigCommand{r: r}
}

func (c *ConfigCommand) Handle(ctx *bot.InteractionContext) error {
	if ctx.IsAutocomplete() {
		return c.handleAutocomplete(ctx)
	}

	i := ctx.Interaction()
	options := ctx.Options()

	switch options[0].Name {
	case "timezone":
		timezone := discordutil.GetRequiredStringOption(options, "timezone")

		if !isValidTimezone(timezone) {
			return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
				Content: "Invalid timezone",
			})
		}

		err := c.r.SetUserTimezone(ctx.Context(), discordutil.GetInteractionUser(i).ID, timezone)
		if err != nil {
			return err
		}

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Timezone set!",
		})
	}
	return nil
}

func (c *ConfigCommand) handleAutocomplete(ctx *bot.InteractionContext) error {
	focuedOption := discordutil.GetFocusedOption(ctx.Options())

	if focuedOption == nil {
		return nil
	}

	switch focuedOption.Name {
	case "timezone":
		const maxResults = 25
		timezone := focuedOption.StringValue()
		results := make([]*discordgo.ApplicationCommandOptionChoice, 0, maxResults)

		if timezone == "" {
			for i, tz := range validTimezones {
				if i >= maxResults {
					break
				}

				results = append(results, &discordgo.ApplicationCommandOptionChoice{
					Name:  tz,
					Value: tz,
				})
			}

		} else {
			for _, tz := range validTimezones {
				target := getComparableTimezoneString(timezone)
				compare := getComparableTimezoneString(tz)

				if strings.Contains(compare, target) {
					results = append(results, &discordgo.ApplicationCommandOptionChoice{
						Name:  tz,
						Value: tz,
					})
				}

				if len(results) >= maxResults {
					break
				}
			}
		}

		return ctx.Respond(discordgo.InteractionApplicationCommandAutocompleteResult, &discordgo.InteractionResponseData{
			Choices: results,
		})
	default:
		return nil
	}
}

func getComparableTimezoneString(tzString string) string {
	tzStr := strings.Builder{}
	for _, c := range tzString {
		if strings.ContainsRune(" \t_/", c) {
			continue
		}

		tzStr.WriteRune(unicode.ToLower(c))
	}

	return tzStr.String()
}
