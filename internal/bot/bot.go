package bot

import (
	"context"
	"log"

	"github.com/UTD-JLA/botsu/internal/guilds"
	"github.com/bwmarrin/discordgo"
)

var unexpectedErrorMessage = &discordgo.WebhookParams{
	Content: "An unexpected error occurred!",
}

type Bot struct {
	session         *discordgo.Session
	commands        CommandCollection
	createdCommands []*discordgo.ApplicationCommand
	destroyOnClose  bool
	guildRepo       *guilds.GuildRepository
}

func NewBot(guildRepo *guilds.GuildRepository) *Bot {
	return &Bot{
		commands:  NewCommandCollection(),
		guildRepo: guildRepo,
	}
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Bot is ready")
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand && i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	ctx := NewInteractionContext(s, i, context.Background())

	defer ctx.Cancel()

	err := b.commands.Handle(ctx)
	if err != nil {
		log.Println("Error handling command", err)

		// if this is a command, and we haven't responded yet, respond with an error
		if ctx.IsCommand() && ctx.Deferred() {
			_, err = ctx.RespondOrFollowup(unexpectedErrorMessage, false)

			if err != nil {
				log.Println("Failed to respond to command", err)
			}
		}
	}
}

func (b *Bot) onMemberRemove(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	err := b.guildRepo.RemoveMembers(context.Background(), m.GuildID, []string{m.User.ID})
	if err != nil {
		log.Println("Failed to delete guild member", err)
	}
}

func (b *Bot) SetDestroyCommandsOnClose(destroy bool) {
	b.destroyOnClose = destroy
}

func (b *Bot) AddCommand(data *discordgo.ApplicationCommand, cmd CommandHandler) {
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

	err = s.Open()
	if err != nil {
		return err
	}

	b.session = s

	log.Println("Creating commands")

	for _, cmd := range b.commands {
		// if len(cmd.Data.Options) > 0 {
		// 	fmt.Printf("%v\n", ref.DerefArray(cmd.Data.Options[0].Options))
		// }
		cmd, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, "", cmd.Data)
		if err != nil {
			return err
		}
		b.createdCommands = append(b.createdCommands, cmd)
	}

	return nil
}

func (b *Bot) Close() {
	if b.destroyOnClose {
		for _, c := range b.createdCommands {
			err := b.session.ApplicationCommandDelete(b.session.State.User.ID, "", c.ID)
			if err != nil {
				log.Println("Failed to delete command", c.Name, err)
			}
		}
	}

	b.session.Close()
}
