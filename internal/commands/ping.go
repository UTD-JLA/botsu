package commands

import (
	"errors"
	"time"

	"github.com/UTD-JLA/botsu/pkg/discordutil"

	"github.com/bwmarrin/discordgo"
)

var PingCommandData = &discordgo.ApplicationCommand{
	Name:        "ping",
	Description: "Ping!",
}

type PingCommand struct{}

func NewPingCommand() *PingCommand {
	return &PingCommand{}
}

func (c *PingCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return errors.New("invalid interaction type")
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Ping!",
							Style:    discordgo.SuccessButton,
							CustomID: "ping",
						},
					},
				},
			},
		},
	})

	if err != nil {
		return err
	}

	msg, err := s.InteractionResponse(i.Interaction)

	if err != nil {
		return err
	}

	// we are now executing an application command
	componentCollector := discordutil.NewMessageComponentCollector(s)
	defer componentCollector.Close()

	// start collecting interactions
	componentCollector.Start(func(ci *discordgo.InteractionCreate) bool {
		return ci.Message.ID == msg.ID &&
			ci.MessageComponentData().CustomID == "ping" &&
			discordutil.IsSameInteractionUser(ci, i)
	})

	componentInteractions := componentCollector.Channel()
	timeout := time.After(time.Second * 20)

loop:
	for x := 0; x < 5; x++ {
		select {
		case ci := <-componentInteractions:
			var componentResponse discordgo.InteractionResponseData

			componentResponse = discordgo.InteractionResponseData{
				Content: "Pong!",
			}

			err = s.InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &componentResponse,
			})

			if err != nil {
				return err
			}
		case <-timeout:
			break loop
		}
	}

	content := "Pong!"

	// remove button from message
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &[]discordgo.MessageComponent{},
	})

	return err

	// wait for an interaction
	// ci, err := componentCollector.NextInteraction(time.Second * 10)

	// var componentResponse discordgo.InteractionResponseData

	// if err != nil {
	// 	componentResponse = discordgo.InteractionResponseData{
	// 		Content: "Timed out!",
	// 	}
	// } else {
	// 	componentResponse = discordgo.InteractionResponseData{
	// 		Content: "Pong!",
	// 	}
	// }

	// return s.InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
	// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
	// 	Data: &componentResponse,
	// })
}
