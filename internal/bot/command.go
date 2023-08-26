package bot

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler interface {
	HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

type Command struct {
	Handler CommandHandler
	Data    *discordgo.ApplicationCommand
}

type CommandCollection map[string]Command

func NewCommandCollection() CommandCollection {
	return CommandCollection{}
}

func (c CommandCollection) Add(data *discordgo.ApplicationCommand, handler CommandHandler) {
	c[data.Name] = Command{
		Handler: handler,
		Data:    data,
	}
}

func (c CommandCollection) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	cmd, ok := c[i.ApplicationCommandData().Name]
	if !ok {
		return nil
	}

	return cmd.Handler.HandleInteraction(s, i)
}
