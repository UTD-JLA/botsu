package bot

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/UTD-JLA/botsu/internal/guilds"
	"github.com/bwmarrin/discordgo"
)

var unexpectedErrorMessage = &discordgo.WebhookParams{
	Content: "An unexpected error occurred!",
}

type Bot struct {
	logger          *slog.Logger
	session         *discordgo.Session
	commands        CommandCollection
	createdCommands []*discordgo.ApplicationCommand
	destroyOnClose  bool
	guildRepo       *guilds.GuildRepository
	noPanic         bool
}

func NewBot(logger *slog.Logger, guildRepo *guilds.GuildRepository) *Bot {
	return &Bot{
		logger:    logger,
		commands:  NewCommandCollection(),
		guildRepo: guildRepo,
	}
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.logger.Info("Bot is ready", slog.String("user", r.User.String()))
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.logger.Debug(
		"Interaction received",
		slog.String("interaction", i.Interaction.ID),
		slog.String("user", i.Interaction.User.String()),
		slog.String("guild", i.Interaction.GuildID),
		slog.String("type", i.Type.String()),
	)
	if i.Type != discordgo.InteractionApplicationCommand && i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	subLogger := b.logger.
		WithGroup("interaction").
		With(slog.String("id", i.Interaction.ID)).
		With(slog.String("user", i.Interaction.User.String())).
		With(slog.String("guild", i.Interaction.GuildID)).
		With(slog.String("type", i.Type.String())).
		With(slog.String("command", i.ApplicationCommandData().Name)).
		WithGroup("handler")

	ctx := NewInteractionContext(subLogger, s, i, context.Background())

	defer ctx.Cancel()

	if b.noPanic {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				ctx.Logger.Error("Panic occurred", slog.Any("panic", r), slog.Any("stack", string(stack)))
				_, err := ctx.RespondOrFollowup(unexpectedErrorMessage, false)

				if err != nil {
					ctx.Logger.Error("Failed to send error message", slog.String("err", err.Error()))
				}
			}
		}()
	}

	err := b.commands.Handle(ctx)
	if err != nil {
		ctx.Logger.Error("Failed to handle interaction", slog.String("err", err.Error()))

		if ctx.IsCommand() {
			_, err = ctx.RespondOrFollowup(unexpectedErrorMessage, false)

			if err != nil {
				ctx.Logger.Error("Failed to send error message", slog.String("err", err.Error()))
			}
		}
	}
}

func (b *Bot) onMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	b.logger.Debug("Member left", slog.String("guild", m.GuildID), slog.String("user", m.User.String()))
	err := b.guildRepo.RemoveMembers(context.Background(), m.GuildID, []string{m.User.ID})
	if err != nil {
		b.logger.Error("Failed to remove member", slog.String("err", err.Error()))
	}
}

func (b *Bot) SetNoPanic(noPanic bool) {
	b.logger.Debug("Setting no panic", slog.Bool("no_panic", noPanic))
	b.noPanic = noPanic
}

func (b *Bot) SetDestroyCommandsOnClose(destroy bool) {
	b.logger.Debug("Setting destroy commands on close", slog.Bool("destroy", destroy))
	b.destroyOnClose = destroy
}

func (b *Bot) AddCommand(data *discordgo.ApplicationCommand, cmd CommandHandler) {
	b.logger.Debug("Adding command", slog.String("command_name", data.Name))
	b.commands.Add(data, cmd)
}

func (b *Bot) Login(token string, intent discordgo.Intent) error {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	s.AddHandler(b.onReady)
	s.AddHandler(b.onInteractionCreate)
	s.AddHandler(b.onMemberRemove)

	s.Identify.Intents = intent

	if err = s.Open(); err != nil {
		return err
	}

	b.session = s
	b.logger.Info("Creating commands")

	cmdData := make([]*discordgo.ApplicationCommand, 0, len(b.commands))

	for _, cmd := range b.commands {
		cmdData = append(cmdData, cmd.Data)
	}

	b.createdCommands, err = b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", cmdData)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) Close() {
	b.logger.Debug("Close() called")
	defer b.session.Close()

	if b.destroyOnClose {
		b.logger.Debug("Destroying commands")

		_, err := b.session.ApplicationCommandBulkOverwrite(b.session.State.User.ID, "", []*discordgo.ApplicationCommand{})

		if err != nil {
			b.logger.Error("Failed to destroy commands", slog.String("err", err.Error()))
		}
	}
}
