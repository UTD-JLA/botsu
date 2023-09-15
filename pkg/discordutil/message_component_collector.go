package discordutil

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type InteractionFilter func(i *discordgo.InteractionCreate) bool

func CollectComponentInteraction(ctx context.Context, s *discordgo.Session, f InteractionFilter) <-chan *discordgo.InteractionCreate {
	ch := make(chan *discordgo.InteractionCreate)

	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent && f(i) {
			ch <- i
		}
	})

	go func() {
		<-ctx.Done()
		removeHandler()
		close(ch)
	}()

	return ch
}

func AcceptAllInteractionFilter(i *discordgo.InteractionCreate) bool {
	return true
}

func NewUserFilter(userID string) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		return GetInteractionUser(i).ID == userID
	}
}

func NewMessageFilter(messageID string) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		return i.Message.ID == messageID
	}
}

func NewMultiFilter(filters ...InteractionFilter) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		for _, f := range filters {
			if !f(i) {
				return false
			}
		}
		return true
	}
}
