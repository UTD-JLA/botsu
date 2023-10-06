package commands

import (
	"fmt"
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
		{
			Name:        "vn-speed",
			Description: "Set your VN reading speed (char/min)",
			Type:        discordgo.ApplicationCommandOptionNumber,
			Required:    false,
		},
		{
			Name:        "book-speed",
			Description: "Set your book reading speed (page/min)",
			Type:        discordgo.ApplicationCommandOptionNumber,
			Required:    false,
		},
		{
			Name:        "manga-speed",
			Description: "Set your manga reading speed (page/min)",
			Type:        discordgo.ApplicationCommandOptionNumber,
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

func (c *ConfigCommand) Handle(ctx *bot.InteractionContext) error {
	if ctx.IsAutocomplete() {
		return c.handleAutocomplete(ctx)
	}

	i := ctx.Interaction()
	options := ctx.Options()

	if len(options) != 1 {
		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "You must provide one option!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	switch options[0].Name {
	case "timezone":
		timezone := discordutil.GetRequiredStringOption(options, "timezone")

		if !isValidTimezone(timezone) {
			return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
				Content: "Invalid timezone",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
		}

		err := c.r.SetUserTimezone(ctx.Context(), discordutil.GetInteractionUser(i).ID, timezone)
		if err != nil {
			return err
		}

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Timezone set!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	case "vn-speed":
		vnSpeed := float32(discordutil.GetRequiredFloatOption(options, "vn-speed"))
		err := c.r.SetVisualNovelReadingSpeed(ctx.Context(), discordutil.GetInteractionUser(i).ID, vnSpeed)
		if err != nil {
			return err
		}

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Visual novel reading speed set!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	case "book-speed":
		bookSpeed := float32(discordutil.GetRequiredFloatOption(options, "book-speed"))
		err := c.r.SetBookReadingSpeed(ctx.Context(), discordutil.GetInteractionUser(i).ID, bookSpeed)
		if err != nil {
			return err
		}

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Book reading speed set!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	case "manga-speed":
		mangaSpeed := float32(discordutil.GetRequiredFloatOption(options, "manga-speed"))
		err := c.r.SetBookReadingSpeed(ctx.Context(), discordutil.GetInteractionUser(i).ID, mangaSpeed)
		if err != nil {
			return err
		}

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Manga reading speed set!",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
	}

	return fmt.Errorf("unexpected option: %s", options[0].Name)
}

func (c *ConfigCommand) handleAutocomplete(ctx *bot.InteractionContext) error {
	focusedOption := discordutil.GetFocusedOption(ctx.Options())

	if focusedOption == nil {
		return nil
	}

	switch focusedOption.Name {
	case "timezone":
		const maxResults = 25
		timezone := focusedOption.StringValue()
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
