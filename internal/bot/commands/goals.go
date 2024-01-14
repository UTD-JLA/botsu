package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/goals"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
	"github.com/hashicorp/cronexpr"
)

var GoalCommandData = &discordgo.ApplicationCommand{
	Name:        "goal",
	Description: "Manage your goals.",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create a new goal.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "The name of the goal.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "target",
					Description: "The target duration of the goal.",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "cron",
					Description: "The cron expression for the goal.",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Daily",
							Value: "@daily",
						},
						{
							Name:  "Weekly",
							Value: "@weekly",
						},
						{
							Name:  "Monthly",
							Value: "@monthly",
						},
						{
							Name:  "Yearly",
							Value: "@yearly",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "activity-type",
					Description: "The type of activity to track.",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Listening",
							Value: activities.ActivityImmersionTypeListening,
						},
						{
							Name:  "Reading",
							Value: activities.ActivityImmersionTypeReading,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "media-type",
					Description: "The type of media to track.",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Visual Novel",
							Value: activities.ActivityMediaTypeVisualNovel,
						},
						{
							Name:  "Book",
							Value: activities.ActivityMediaTypeBook,
						},
						{
							Name:  "Manga",
							Value: activities.ActivityMediaTypeManga,
						},
						{
							Name:  "Anime",
							Value: activities.ActivityMediaTypeAnime,
						},
						{
							Name:  "Video",
							Value: activities.ActivityMediaTypeVideo,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "youtube-channels",
					Description: "The YouTube channels to track (comma separated, e.g. @HakuiKoyori,@ui_shig,@MinatoAqua).",
					Required:    false,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List your goals.",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "delete",
			Description: "Delete a goal.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "id",
					Description: "The ID of the goal.",
					Required:    true,
				},
			},
		},
	},
}

type GoalCommand struct {
	goals *goals.GoalService
}

func NewGoalCommand(goals *goals.GoalService) *GoalCommand {
	return &GoalCommand{goals: goals}
}

func (c *GoalCommand) Handle(cmd *bot.InteractionContext) error {
	if len(cmd.Options()) == 0 {
		return bot.ErrInvalidOptions
	}

	subcommand := cmd.Options()[0]

	switch subcommand.Name {
	case "create":
		return c.handleCreate(cmd, subcommand)
	case "list":
		return c.handleList(cmd, subcommand)
	default:
		return bot.ErrInvalidOptions
	}
}

func (c *GoalCommand) handleList(cmd *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	goals, err := c.goals.FindByUserID(cmd.ResponseContext(), cmd.User().ID)

	if err != nil {
		return fmt.Errorf("failed to find goals: %w", err)
	}

	if len(goals) == 0 {
		_, err := cmd.Followup(
			&discordgo.WebhookParams{
				Content: "You have no goals!",
			},
			false,
		)

		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Goals").
		SetColor(discordutil.ColorPrimary).
		SetTimestamp(time.Now())

	// pageFunc := func(i int) *discordgo.MessageEmbed {
	// 	embed.ClearFields()
	// 	embed.SetFooter(fmt.Sprintf("Page %d/%d", i+1, len(goals)))
	// 	embed.AddField("Goal", fmt.Sprintf("%s (%d)", goals[i].Name, goals[i].ID), false)
	// 	embed.AddField("Target", goals[i].Target.String(), true)
	// 	embed.AddField("Current", goals[i].Current.String(), true)
	// 	embed.AddField("Due At", fmt.Sprintf("<t:%d>", goals[i].DueAt.Unix()), true)
	// 	embed.AddField("Cron", goals[i].Cron, true)

	// 	if goals[i].ActivityType != nil {
	// 		embed.AddField("Activity Type", *goals[i].ActivityType, true)
	// 	}

	// 	if goals[i].MediaType != nil {
	// 		embed.AddField("Media Type", *goals[i].MediaType, true)
	// 	}

	// 	if len(goals[i].YoutubeChannels) > 0 {
	// 		embed.AddField("YouTube Channels", strings.Join(goals[i].YoutubeChannels, ", "), true)
	// 	}

	// 	return embed.MessageEmbed
	// }

	for _, goal := range goals {
		nextDueDate, err := c.goals.NextCron(cmd.ResponseContext(), goal)

		if err != nil {
			return fmt.Errorf("failed to calculate next due date: %w", err)
		}

		title := fmt.Sprintf("%s (%d)", goal.Name, goal.ID)

		embed.AddField(title, fmt.Sprintf(
			"Progress: %s / %s **(%.2f%%)**\nNext Reset: <t:%d>",
			goal.Current,
			goal.Target,
			goal.Current.Seconds()/goal.Target.Seconds()*100,
			nextDueDate.Unix(),
		), false)
	}

	return cmd.Respond(
		discordgo.InteractionResponseChannelMessageWithSource,
		&discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed.MessageEmbed},
		},
	)
}

func (c *GoalCommand) handleCreate(cmd *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	name, err := discordutil.GetRequiredStringOption(subcommand.Options, "name")
	if err != nil {
		return err
	}

	target, err := discordutil.GetRequiredIntOption(subcommand.Options, "target")
	if err != nil {
		return err
	}

	cron, err := discordutil.GetRequiredStringOption(subcommand.Options, "cron")
	if err != nil {
		return err
	}

	if _, err = cronexpr.Parse(cron); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	activityType := discordutil.GetStringOption(subcommand.Options, "activity-type")
	mediaType := discordutil.GetStringOption(subcommand.Options, "media-type")
	ytChannels := discordutil.GetStringOption(subcommand.Options, "youtube-channels")

	goal := &goals.Goal{}

	if activityType != nil {
		goal.ActivityType = activityType
	}

	if mediaType != nil {
		goal.MediaType = mediaType
	}

	if ytChannels != nil {
		goal.YoutubeChannels = strings.Split(*ytChannels, ",")
	}

	goal.Name = name
	goal.Target = time.Duration(target) * time.Minute
	goal.Cron = cron
	goal.UserID = cmd.User().ID
	goal.DueAt, err = c.goals.NextCron(cmd.ResponseContext(), goal)

	if err != nil {
		return fmt.Errorf("failed to calculate due date: %w", err)
	}

	// for testing
	fmt.Printf("goal: %+v\n", goal)

	if err := c.goals.Create(cmd.ResponseContext(), goal); err != nil {
		return fmt.Errorf("failed to create goal: %w", err)
	}

	return cmd.Respond(
		discordgo.InteractionResponseChannelMessageWithSource,
		&discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Goal **%s** created!", goal.Name),
		},
	)
}
