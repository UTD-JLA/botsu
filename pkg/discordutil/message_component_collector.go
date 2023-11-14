package discordutil

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type MessageComponentCollector struct {
	channels      map[string]chan *discordgo.InteractionCreate
	removeHandler func()
	mu            sync.Mutex
}

func NewMessageComponentCollector(s *discordgo.Session) *MessageComponentCollector {
	cc := &MessageComponentCollector{
		channels: make(map[string]chan *discordgo.InteractionCreate),
		mu:       sync.Mutex{},
	}

	cc.removeHandler = s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			cc.mu.Lock()
			defer cc.mu.Unlock()
			if ch, ok := cc.channels[i.Message.ID]; ok {
				ch <- i
			}
		}
	})

	return cc
}

func (cc *MessageComponentCollector) Close() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.removeHandler()

	for id, ch := range cc.channels {
		close(ch)
		delete(cc.channels, id)
	}
}

func (cc *MessageComponentCollector) Collect(ctx context.Context, messageID string, f InteractionFilter) <-chan *discordgo.InteractionCreate {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	ch := make(chan *discordgo.InteractionCreate)
	cc.channels[messageID] = ch

	go func() {
		<-ctx.Done()
		cc.mu.Lock()
		if ch, ok := cc.channels[messageID]; ok {
			close(ch)
			delete(cc.channels, messageID)
		}
		cc.mu.Unlock()

	}()

	return ch
}

type InteractionFilter func(i *discordgo.InteractionCreate) bool

func CollectComponentInteractions(ctx context.Context, s *discordgo.Session, f InteractionFilter) <-chan *discordgo.InteractionCreate {
	// this channel closes
	ch := make(chan *discordgo.InteractionCreate)
	// this channel does not close
	middleCh := make(chan *discordgo.InteractionCreate)

	removeHandler := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent && f(i) {
			middleCh <- i
		}
	})

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case i := <-middleCh:
				ch <- i
			}
		}

		close(ch)
		removeHandler()
	}()

	return ch
}

func CollectSingleComponentInteraction(ctx context.Context, s *discordgo.Session, f InteractionFilter) (*discordgo.InteractionCreate, error) {
	ch := make(chan *discordgo.InteractionCreate)

	removeFunc := s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent && f(i) {
			ch <- i
		}
	})

	defer removeFunc()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case i := <-ch:
		return i, nil
	}
}

func AcceptAllInteractionFilter(i *discordgo.InteractionCreate) bool {
	return true
}

func NewUserFilter(userID string) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		return GetInteractionUser(i).ID == userID
	}
}

func NewInteractionUserFilter(interaction *discordgo.InteractionCreate) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		return GetInteractionUser(interaction).ID == GetInteractionUser(i).ID
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

func NewCustomIDFilter(customID string) InteractionFilter {
	return func(i *discordgo.InteractionCreate) bool {
		return i.MessageComponentData().CustomID == customID
	}
}
