package discordutil

import (
	"errors"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type InteractionFilter func(i *discordgo.InteractionCreate) bool

type MessageComponentCollector struct {
	ch            chan *discordgo.InteractionCreate
	session       *discordgo.Session
	removeHandler func()
	closedMu      sync.RWMutex
	closed        bool
}

func NewMessageComponentCollector(s *discordgo.Session) *MessageComponentCollector {
	return &MessageComponentCollector{
		ch:            make(chan *discordgo.InteractionCreate),
		session:       s,
		removeHandler: func() {},
		closed:        false,
	}
}

func (c *MessageComponentCollector) Channel() <-chan *discordgo.InteractionCreate {
	return c.ch
}

func (c *MessageComponentCollector) Start(f InteractionFilter) {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closed {
		return
	}

	c.removeHandler = c.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// it is okay for multiple goroutines to write to the channel
		// as long as we are not closing it
		c.closedMu.RLock()
		defer c.closedMu.RUnlock()

		if c.closed {
			return
		}

		if i.Type == discordgo.InteractionMessageComponent && f(i) {
			c.ch <- i
		}
	})
}

func (c *MessageComponentCollector) NextInteraction(timeout time.Duration) (*discordgo.InteractionCreate, error) {
	select {
	case i := <-c.ch:
		return i, nil
	case <-time.After(timeout):
		return nil, errors.New("timeout")
	}
}

func (c *MessageComponentCollector) Close() {
	c.closedMu.Lock()
	defer c.closedMu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	c.removeHandler()
	close(c.ch)
}
