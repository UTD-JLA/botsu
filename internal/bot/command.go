package bot

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHandler interface {
	Handle(ctx *InteractionContext) error
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

func (c CommandCollection) Handle(ctx *InteractionContext) error {
	cmd, ok := c[ctx.Interaction().ApplicationCommandData().Name]
	if !ok {
		return nil
	}

	return cmd.Handler.Handle(ctx)
}
