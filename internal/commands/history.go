package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

var HistoryCommandData = &discordgo.ApplicationCommand{
	Name:        "history",
	Description: "View your activity history",
}

type HistoryCommand struct {
	r *activities.ActivityRepository
}

func NewHistoryCommand(r *activities.ActivityRepository) *HistoryCommand {
	return &HistoryCommand{r: r}
}

func (c *HistoryCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	offset := 0
	user := discordutil.GetInteractionUser(i)
	page, err := c.r.PageByUserID(context.Background(), user.ID, 6, offset)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity History").
		SetDescription("Here are your last 10 activities!").
		SetColor(discordutil.ColorPrimary).
		SetFooter(fmt.Sprintf("Page %d of %d", page.Page, page.PageCount), "")

	for _, activity := range page.Activities {
		embed.AddField(activity.Date.Format(time.DateTime), activity.Name, true)
	}

	nextButton := discordgo.Button{
		Label:    "Next",
		Style:    discordgo.SuccessButton,
		CustomID: "history_next",
		Disabled: page.Page == page.PageCount,
	}

	previousButton := discordgo.Button{
		Label:    "Previous",
		Style:    discordgo.SuccessButton,
		CustomID: "history_previous",
		Disabled: true,
	}

	collector := discordutil.NewMessageComponentCollector(s)
	defer collector.Close()

	msg, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					previousButton,
					nextButton,
				},
			},
		},
	})

	collector.Start(func(ci *discordgo.InteractionCreate) bool {
		return ci.Message.ID == msg.ID &&
			discordutil.IsSameInteractionUser(ci, i)
	})

	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(time.Minute * 3):
		case ci := <-collector.Channel():
			if ci.MessageComponentData().CustomID == "history_previous" {
				offset -= 6
				page, err = c.r.PageByUserID(context.Background(), user.ID, 6, offset)

				if err != nil {
					return err
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
			} else if ci.MessageComponentData().CustomID == "history_next" {
				offset += 6
				page, err = c.r.PageByUserID(context.Background(), user.ID, 6, offset)

				if err != nil {
					return err
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
