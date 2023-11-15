package discordutil

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// TODO: handle this case
var ErrMessageHasHandler = errors.New("the selected message already has an associated message component interaction handler")

type handlerWithFilter struct {
	ch chan *discordgo.InteractionCreate
	f  func(*discordgo.InteractionCreate) bool
}

type MessageComponentCollector struct {
	handlers      map[string]handlerWithFilter
	removeHandler func()
	mu            sync.Mutex
}

func NewMessageComponentCollector(s *discordgo.Session) *MessageComponentCollector {
	cc := &MessageComponentCollector{
		handlers: make(map[string]handlerWithFilter),
		mu:       sync.Mutex{},
	}

	cc.removeHandler = s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			cc.mu.Lock()
			defer cc.mu.Unlock()
			if handler, ok := cc.handlers[i.Message.ID]; ok && handler.f(i) {
				handler.ch <- i
			}
		}
	})

	return cc
}

func (cc *MessageComponentCollector) Close() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.removeHandler()

	for id, handler := range cc.handlers {
		close(handler.ch)
		delete(cc.handlers, id)
	}
}

func (cc *MessageComponentCollector) CollectOnce(
	ctx context.Context,
	messageID string,
	f InteractionFilter,
) (i *discordgo.InteractionCreate, err error) {
	logger := slog.Default().With(slog.String("message.id", messageID))
	logger.Debug("Collecting single component interaction")

	cc.mu.Lock()

	if _, ok := cc.handlers[messageID]; ok {
		logger.Debug("Attempted to create second message component interaction handler")
		cc.mu.Unlock()
		return nil, ErrMessageHasHandler
	}

	ch := make(chan *discordgo.InteractionCreate, 1)
	cc.handlers[messageID] = handlerWithFilter{ch, f}
	cc.mu.Unlock()

	select {
	case <-ctx.Done():
		logger.Debug("Unexpected context done (single)")
		err = ctx.Err()
		return
	case i = <-ch:
		logger.Debug("Received interaction, deleting channel (single)")
		cc.mu.Lock()
		delete(cc.handlers, messageID)
		cc.mu.Unlock()
		return
	}
}

func (cc *MessageComponentCollector) Collect(
	ctx context.Context,
	messageID string,
	f InteractionFilter,
) (<-chan *discordgo.InteractionCreate, error) {
	logger := slog.Default().With(slog.String("message.id", messageID))
	logger.Debug("Collecting component interactions")

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if _, ok := cc.handlers[messageID]; ok {
		logger.Debug("Attempted to create second message component interaction handler")
		return nil, ErrMessageHasHandler
	}

	ch := make(chan *discordgo.InteractionCreate)
	cc.handlers[messageID] = handlerWithFilter{ch, f}

	go func() {
		<-ctx.Done()

		logger.Debug("Context done, removing channel")

		cc.mu.Lock()
		if handler, ok := cc.handlers[messageID]; ok {
			close(handler.ch)
			delete(cc.handlers, messageID)
		}
		cc.mu.Unlock()
	}()

	return ch, nil
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
		defer close(ch)
		defer removeHandler()

		for {
			select {
			case <-ctx.Done():
				return
			case i := <-middleCh:
				ch <- i
			}
		}
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
		return IsSameInteractionUser(i, interaction)
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
