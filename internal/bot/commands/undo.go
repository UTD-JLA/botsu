package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/UTD-JLA/botsu/pkg/ref"
	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v5"
)

var UndoCommandData = &discordgo.ApplicationCommand{
	Name:        "undo",
	Description: "Undo the last activity you logged",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			MinValue:    ref.New(0.0),
			Name:        "id",
			Description: "The ID of the activity to undo",
			Required:    false,
		},
	},
}

type UndoCommand struct {
	r *activities.ActivityRepository
}

func NewUndoCommand(r *activities.ActivityRepository) *UndoCommand {
	return &UndoCommand{r: r}
}

func (c *UndoCommand) Handle(ctx *bot.InteractionContext) error {
	id := discordutil.GetUintOption(ctx.Options(), "id")

	if id == nil {
		return c.undoLastActivity(ctx)
	} else {
		return c.undoActivity(ctx, *id)
	}
}

func (c *UndoCommand) undoActivity(ctx *bot.InteractionContext, id uint64) error {
	activity, err := c.r.GetByID(ctx.ResponseContext(), id, ctx.Interaction().GuildID)

	if err == pgx.ErrNoRows {
		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Activity not found.",
		})
	} else if err != nil {
		return err
	} else if activity.UserID != discordutil.GetInteractionUser(ctx.Interaction()).ID {
		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "You can only undo your own activities!",
		})
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Undo Activity").
		SetDescription("Are you sure you want to undo this activity?").
		AddField("Name", activity.Name, true).
		AddField("Date", fmt.Sprintf("<t:%d>", activity.Date.Unix()), true).
		AddField("Created At", fmt.Sprintf("<t:%d>", activity.CreatedAt.Unix()), true).
		AddField("Duration", activity.Duration.String(), true).
		SetFooter("This cannot be undone!", "").
		SetColor(discordutil.ColorWarning).
		Build()

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Yes",
				Style:    discordgo.DangerButton,
				CustomID: "undo_confirm",
			},
			discordgo.Button{
				Label:    "No",
				Style:    discordgo.SecondaryButton,
				CustomID: "undo_cancel",
			},
		},
	}

	err = ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row},
		Flags:      discordgo.MessageFlagsEphemeral,
	})

	if err != nil {
		return err
	}

	msg, err := ctx.Session().InteractionResponse(ctx.Interaction().Interaction)

	if err != nil {
		return err
	}

	collectionContext, cancel := context.WithTimeout(ctx.Context(), 15*time.Second)
	defer cancel()

	interactions := discordutil.CollectComponentInteraction(collectionContext, ctx.Session(), discordutil.NewMultiFilter(
		discordutil.NewMessageFilter(msg.ID),
		discordutil.NewUserFilter(discordutil.GetInteractionUser(ctx.Interaction()).ID),
	))

	ci, ok := <-interactions

	if !ok {
		_, err := ctx.Session().InteractionResponseEdit(ctx.Interaction().Interaction, &discordgo.WebhookEdit{
			Content:    ref.New("Timed out!"),
			Components: &[]discordgo.MessageComponent{},
			Embeds:     &[]*discordgo.MessageEmbed{},
		})

		return err
	}

	ciCtx, cancel := context.WithDeadline(ctx.Context(), discordutil.GetInteractionResponseDeadline(ci.Interaction))
	defer cancel()

	if ci.MessageComponentData().CustomID == "undo_confirm" {
		err = c.r.DeleteById(ciCtx, activity.ID)

		if err != nil {
			return err
		}

		err := ctx.Session().InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Activity deleted!",
				Components: []discordgo.MessageComponent{},
				Embeds:     []*discordgo.MessageEmbed{},
			},
		})

		return err
	} else if ci.MessageComponentData().CustomID == "undo_cancel" {
		err := ctx.Session().InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Cancelled!",
				Components: []discordgo.MessageComponent{},
				Embeds:     []*discordgo.MessageEmbed{},
			},
		})

		return err
	} else {
		return errors.New("invalid custom id")
	}
}

func (c *UndoCommand) undoLastActivity(ctx *bot.InteractionContext) error {
	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID
	activity, err := c.r.GetLatestByUserID(ctx.ResponseContext(), userID, ctx.Interaction().GuildID)

	if err == pgx.ErrNoRows {
		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "You have no activities to undo.",
		})
	} else if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Undo Activity").
		SetDescription("Are you sure you want to undo this activity?").
		AddField("Name", activity.Name, true).
		AddField("Date", fmt.Sprintf("<t:%d>", activity.Date.Unix()), true).
		AddField("Created At", fmt.Sprintf("<t:%d>", activity.CreatedAt.Unix()), true).
		AddField("Duration", activity.Duration.String(), true).
		SetColor(discordutil.ColorWarning).
		SetFooter("This cannot be undone!", "").
		Build()

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Yes",
				Style:    discordgo.DangerButton,
				CustomID: "undo_confirm",
			},
			discordgo.Button{
				Label:    "No",
				Style:    discordgo.SecondaryButton,
				CustomID: "undo_cancel",
			},
		},
	}

	err = ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row},
		Flags:      discordgo.MessageFlagsEphemeral,
	})

	if err != nil {
		return err
	}

	msg, err := ctx.Session().InteractionResponse(ctx.Interaction().Interaction)

	if err != nil {
		return err
	}

	collectionContext, cancel := context.WithTimeout(ctx.Context(), 15*time.Second)

	defer cancel()

	interactions := discordutil.CollectComponentInteraction(collectionContext, ctx.Session(), discordutil.NewMultiFilter(
		discordutil.NewMessageFilter(msg.ID),
		discordutil.NewUserFilter(userID),
	))

	ci, ok := <-interactions

	if !ok {
		_, err := ctx.Session().InteractionResponseEdit(ctx.Interaction().Interaction, &discordgo.WebhookEdit{
			Content:    ref.New("Timed out!"),
			Components: &[]discordgo.MessageComponent{},
			Embeds:     &[]*discordgo.MessageEmbed{},
		})

		return err
	}

	ciCtx, cancel := context.WithDeadline(ctx.Context(), discordutil.GetInteractionResponseDeadline(ci.Interaction))
	defer cancel()

	if ci.MessageComponentData().CustomID == "undo_confirm" {
		err = c.r.DeleteById(ciCtx, activity.ID)

		if err != nil {
			return err
		}

		err := ctx.Session().InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Activity deleted!",
				Components: []discordgo.MessageComponent{},
				Embeds:     []*discordgo.MessageEmbed{},
			},
		})

		return err
	} else if ci.MessageComponentData().CustomID == "undo_cancel" {
		err := ctx.Session().InteractionRespond(ci.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    "Cancelled!",
				Components: []discordgo.MessageComponent{},
				Embeds:     []*discordgo.MessageEmbed{},
			},
		})

		return err
	} else {
		return errors.New("invalid custom id")
	}
}
