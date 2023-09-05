package commands

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

var ExportCommandData = &discordgo.ApplicationCommand{
	Name:        "export",
	Description: "Export your activities to a JSONL file.",
}

type ExportCommand struct {
	r         *activities.ActivityRepository
	history   map[string]time.Time
	historyMu sync.Mutex
}

func NewExportCommand(r *activities.ActivityRepository) *ExportCommand {
	return &ExportCommand{r: r, history: make(map[string]time.Time)}
}

func (c *ExportCommand) Handle(ctx *bot.InteractionContext) error {
	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	c.historyMu.Lock()

	t, ok := c.history[userID]

	if ok && time.Since(t) < time.Hour*24 {
		c.historyMu.Unlock()

		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "You can only export your activities once every day.",
		})
	}

	c.history[userID] = time.Now()
	c.historyMu.Unlock()

	activities, err := c.r.GetAllByUserID(
		ctx.Context(),
		discordutil.GetInteractionUser(ctx.Interaction()).ID,
		ctx.Interaction().GuildID)

	if err != nil {
		return err
	}

	// create stream to write to
	buffer := bytes.NewBuffer([]byte{})
	stream := bufio.NewWriter(buffer)
	compressed := gzip.NewWriter(stream)

	// write activities to stream
	jsonEncoder := json.NewEncoder(compressed)

	for _, activity := range activities {
		jsonEncoder.Encode(activity)
	}

	// flush stream
	compressed.Close()
	stream.Flush()

	filename := fmt.Sprintf("activities-%s-%s.jsonl.gz", userID, time.Now().Format(time.RFC3339))

	// send file
	return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Files: []*discordgo.File{
			{
				Name:        filename,
				ContentType: "application/gzip",
				Reader:      buffer,
			},
		},
	})
}
