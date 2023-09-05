package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

var HistoryCommandData = &discordgo.ApplicationCommand{
	Name:        "history",
	Description: "View your activity history",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "user",
			Type:        discordgo.ApplicationCommandOptionUser,
			Description: "The user to view the history of. Defaults to yourself.",
			Required:    false,
		},
	},
}

type HistoryCommand struct {
	r *activities.ActivityRepository
}

func NewHistoryCommand(r *activities.ActivityRepository) *HistoryCommand {
	return &HistoryCommand{r: r}
}

func (c *HistoryCommand) Handle(ctx *bot.InteractionContext) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	offset := 0
	i := ctx.Interaction()
	s := ctx.Session()
	user := discordutil.GetUserOption(ctx.Options(), "user", s)

	if user == nil {
		user = discordutil.GetInteractionUser(i)
	}

	page, err := c.r.PageByUserID(ctx.Context(), user.ID, ctx.Interaction().GuildID, 6, offset)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity History").
		SetColor(discordutil.ColorPrimary).
		SetAuthor(user.Username, user.AvatarURL("256"), "").
		SetFooter(fmt.Sprintf("Page %d of %d", page.Page, page.PageCount), "")

	for _, activity := range page.Activities {
		embed.AddField(activity.Date.Format(time.DateTime), activity.Name, true)
	}

	nextButton := discordgo.Button{
		Label:    "Next",
		Style:    discordgo.PrimaryButton,
		CustomID: "history_next",
		Disabled: page.Page == page.PageCount,
	}

	previousButton := discordgo.Button{
		Label:    "Previous",
		Style:    discordgo.SecondaryButton,
		CustomID: "history_previous",
		Disabled: true,
	}

	collector := discordutil.NewMessageComponentCollector(s)
	defer collector.Close()

	msg, err := ctx.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					previousButton,
					nextButton,
				},
			},
		},
	}, true)

	collector.Start(func(ci *discordgo.InteractionCreate) bool {
		return ci.Message.ID == msg.ID &&
			discordutil.IsSameInteractionUser(ci, i)
	})

	if err != nil {
		return err
	}

	timeout := time.After(time.Minute * 3)

	for {
		select {
		case <-timeout:
			return nil
		case ci := <-collector.Channel():
			ciContext, cancel := context.WithDeadline(ctx.Context(), discordutil.GetInteractionResponseDeadline(ci.Interaction))
			defer cancel()

			if ci.MessageComponentData().CustomID == "history_previous" {
				offset -= 6
				page, err = c.r.PageByUserID(ciContext, user.ID, ctx.Interaction().GuildID, 6, offset)

				if err != nil {
					return err
				}

				embed.SetFooter(fmt.Sprintf("Page %d of %d", page.Page, page.PageCount), "")
				embed.ClearFields()

				if page.Page%2 == 0 {
					embed.SetColor(discordutil.ColorSecondary)
				} else {
					embed.SetColor(discordutil.ColorPrimary)
				}

				for _, activity := range page.Activities {
					embed.AddField(activity.Date.Format(time.DateTime), activity.Name, true)
				}

				previousButton.Disabled = page.Page == 1
				nextButton.Disabled = page.Page == page.PageCount

				err := s.InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed.Build()},
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									previousButton,
									nextButton,
								},
							},
						},
					},
				})

				if err != nil {
					return err
				}
			} else if ci.MessageComponentData().CustomID == "history_next" {
				offset += 6
				page, err = c.r.PageByUserID(ciContext, user.ID, ctx.Interaction().GuildID, 6, offset)

				if err != nil {
					return err
				}

				if page.Page%2 == 0 {
					embed.SetColor(discordutil.ColorSecondary)
				} else {
					embed.SetColor(discordutil.ColorPrimary)
				}

				embed.SetFooter(fmt.Sprintf("Page %d of %d", page.Page, page.PageCount), "")
				embed.ClearFields()

				for _, activity := range page.Activities {
					embed.AddField(activity.Date.Format(time.DateTime), activity.Name, true)
				}

				previousButton.Disabled = page.Page == 1
				nextButton.Disabled = page.Page == page.PageCount

				err := s.InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed.Build()},
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									previousButton,
									nextButton,
								},
							},
						},
					},
				})

				if err != nil {
					return err
				}
			}
		}
	}
}
