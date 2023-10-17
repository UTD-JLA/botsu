package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	activitiesPub "github.com/UTD-JLA/botsu/pkg/activities"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
)

var ImportCommandData = &discordgo.ApplicationCommand{
	Name:        "import",
	Description: "Import new data and manage your previous imports",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "botsu-file",
			Description: "Import new data",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file",
					Description: "Import your data from a file exported by Botsu",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List your previous imports",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "undo",
			Description: "Undo an import by timestamp",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timestamp",
					Description: "The timestamp of the import to undo (see `/import list`)",
					Required:    true,
				},
			},
		},
	},
}

type ImportCommand struct {
	r *activities.ActivityRepository
}

func NewImportCommand(r *activities.ActivityRepository) *ImportCommand {
	return &ImportCommand{r}
}

func (c *ImportCommand) handleList(
	ctx context.Context,
	cmd *bot.InteractionContext,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	history, err := c.r.GetRecentImportsByUserID(ctx, cmd.User().ID, 10)

	if err != nil {
		return err
	}

	if len(history) == 0 {
		_, err = cmd.Followup(&discordgo.WebhookParams{
			Content: "No imports found.",
		}, false)

		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Recent Imports").
		SetColor(discordutil.ColorInfo).
		SetTimestamp(time.Now())

	description := strings.Builder{}

	for i, h := range history {
		line := fmt.Sprintf(
			"<t:%d:R> - %d activities: `%d`",
			h.Timestamp.Unix(),
			h.Count,
			h.Timestamp.UnixNano(),
		)

		description.WriteString(line)

		if i != len(history)-1 {
			description.WriteString("\n")
		}
	}

	embed.SetDescription(description.String())

	_, err = cmd.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	}, false)

	return err
}

func (c *ImportCommand) handleUndo(
	ctx context.Context,
	cmd *bot.InteractionContext,
	opts []*discordgo.ApplicationCommandInteractionDataOption,
) error {
	// use string instead of int because it is too large for integer options
	timestampString := discordutil.GetRequiredStringOption(opts, "timestamp")
	timestamp, err := strconv.ParseInt(timestampString, 10, 64)

	if err != nil {
		_, err = cmd.Followup(&discordgo.WebhookParams{
			Content: "Invalid timestamp!",
		}, false)

		return err
	}

	var n int64

	fmt.Printf("Undoing import for user %s at timestamp %s\n", cmd.User().ID, time.Unix(0, timestamp).Format(time.RFC3339Nano))

	if n, err = c.r.UndoImportByUserIDAndTimestamp(ctx, cmd.User().ID, time.Unix(0, timestamp)); err != nil {
		_, err = cmd.Followup(&discordgo.WebhookParams{
			Content: "Failed to undo import!",
		}, false)

		return err
	}

	if n == 0 {
		_, err = cmd.Followup(&discordgo.WebhookParams{
			Content: "No activities were removed.",
		}, false)

		return err
	}

	_, err = cmd.Followup(&discordgo.WebhookParams{
		Content: fmt.Sprintf("Successfully undone import! %d activities were removed.", n),
	}, false)

	return err
}

func (c *ImportCommand) Handle(cmd *bot.InteractionContext) error {
	if err := cmd.DeferResponse(); err != nil {
		return err
	}

	ctx := cmd.Context()
	opts := cmd.Options()
	subcommand := opts[0].Name
	opts = opts[0].Options

	if subcommand == "list" {
		return c.handleList(ctx, cmd, opts)
	}

	if subcommand == "undo" {
		return c.handleUndo(ctx, cmd, opts)
	}

	attachmentID, ok := discordutil.GetRequiredOption(opts, "file").Value.(string)

	if !ok {
		return errors.New("expected string value from attachment option")
	}

	attachment := cmd.Data().Resolved.Attachments[attachmentID]
	extension := strings.ToLower(path.Ext(attachment.Filename))

	if extension != ".gz" && extension != ".jsonl" {
		_, err := cmd.Followup(&discordgo.WebhookParams{
			Content: "Invalid file type!",
		}, false)
		return err
	}

	req, err := http.NewRequest(http.MethodGet, attachment.URL, nil)

	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var as []*activities.Activity

	switch extension {
	case ".gz":
		as, err = activitiesPub.ReadCompressedJSONL(resp.Body)
	case ".jsonl":
		as, err = activitiesPub.ReadJSONL(resp.Body)
	}

	if err != nil {
		return err
	}

	if err := c.r.ImportMany(ctx, as); err != nil {
		log.Println(err)

		_, err = cmd.Followup(&discordgo.WebhookParams{
			Content: "Data was not fully imported!",
		}, false)

		return err
	}

	_, err = cmd.Followup(&discordgo.WebhookParams{
		Content: "Done!",
	}, false)

	return err
}
