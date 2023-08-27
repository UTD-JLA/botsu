package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session         *discordgo.Session
	commands        CommandCollection
	createdCommands []*discordgo.ApplicationCommand
	destroyOnClose  bool
}

func NewBot() *Bot {
	return &Bot{
		commands: NewCommandCollection(),
	}
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Println("Bot is ready")
}

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		fallthrough
	case discordgo.InteractionApplicationCommandAutocomplete:
		err = b.commands.HandleInteraction(s, i)
	}

	if err != nil {
		log.Println("Error executing command:", err)
	}
}

func (b *Bot) SetDestroyCommandsOnClose(destroy bool) {
	b.destroyOnClose = destroy
}

func (b *Bot) AddCommand(data *discordgo.ApplicationCommand, cmd CommandHandler) {
	b.commands.Add(data, cmd)
}

func (b *Bot) Login(token string) error {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return err
	}

	s.AddHandler(b.onReady)
	s.AddHandler(b.onInteractionCreate)

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