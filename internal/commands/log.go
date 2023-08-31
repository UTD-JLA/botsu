package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/UTD-JLA/botsu/pkg/ref"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-module/carbon/v2"
	"github.com/kkdai/youtube/v2"
)

var manualCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:        "name",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Title/name of the activity completed",
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "type",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Type of activity (listening/reading)",
		Required:    true,
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
		Name:        "duration",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Duration spent on the activity",
		MinValue:    ref.New(0.0),
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "date",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Date of activity completion (default is current time)",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "media-type",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Type of media of the activity",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
		Choices: []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  "Anime",
				Value: activities.ActivityMediaTypeAnime,
			},
			{
				Name:  "Manga",
				Value: activities.ActivityMediaTypeManga,
			},
			{
				Name:  "Book",
				Value: activities.ActivityMediaTypeBook,
			},
			{
				Name:  "Video",
				Value: activities.ActivityMediaTypeVideo,
			},
			{
				Name:  "Visual Novel",
				Value: activities.ActivityMediaTypeVisualNovel,
			},
		},
	},
}

var videoCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:        "url",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "URL of the video.",
		Required:    true,
	},
	{
		Name:        "date",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Date of activity completion (default is current time)",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "duration",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Duration spent on the activity",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

var vnCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:        "name",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Title/name of the book read",
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "characters",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Number of characters read (or 0 if unknown)",
		MinValue:    ref.New(0.0),
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "duration",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "How long it took to read (mins, overrides reading-speed)",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "reading-speed",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "How many characters per minute you read (default 150)",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "reading-speed-hourly",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "How many characters per hour you read (overrides reading-speed)",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "date",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Date of activity completion (default is current time)",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

var bookCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:        "name",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Title/name of the book read",
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "pages",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Number of pages read (or 0 if unknown)",
		MinValue:    ref.New(0.0),
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "duration",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "How long it took to read (mins, overrides reading-speed)",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "date",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Date of activity completion (default is current time)",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

var animeCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:        "name",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Title/name of the anime watched",
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "episodes",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Number of episodes watched",
		MinValue:    ref.New(0.0),
		Required:    true,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "episode-duration",
		Type:        discordgo.ApplicationCommandOptionInteger,
		Description: "Duration of each episode (mins, default 24)",
		MinValue:    ref.New(0.0),
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
	{
		Name:        "date",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Date of activity completion (default is current time)",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

var LogCommandData = &discordgo.ApplicationCommand{
	Name:        "log",
	Description: "Log your time spent on language immersion",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "manual",
			Description: "Manually log your immersion time",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     manualCommandOptions,
		},
		{
			Name:        "video",
			Description: "Quickly log a video you watched",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     videoCommandOptions,
		},
		{
			Name:        "vn",
			Description: "Log a visual novel you read",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     vnCommandOptions,
		},
		{
			Name:        "book",
			Description: "Log a book you read",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     bookCommandOptions,
		},
		{
			Name:        "manga",
			Description: "Log a manga you read",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     bookCommandOptions,
		},
		{
			Name:        "anime",
			Description: "Log an anime you watched",
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Options:     animeCommandOptions,
		},
	},
}

type LogCommand struct {
	activityRepo *activities.ActivityRepository
	//userRepo     *users.UserRepository
	userState *users.UserState
	ytClient  youtube.Client
}

func NewLogCommand(ar *activities.ActivityRepository, ur *users.UserRepository) *LogCommand {
	return &LogCommand{
		activityRepo: ar,
		//userRepo:     ur,
		ytClient:  youtube.Client{},
		userState: users.NewUserState(ur),
	}
}

func (c *LogCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	subcommand := i.ApplicationCommandData().Options[0]

	switch subcommand.Name {
	case "manual":
		return c.handleManual(s, i, subcommand)
	case "video":
		return c.handleVideo(s, i, subcommand)
	case "vn":
		return c.handleVisualNovel(s, i, subcommand)
	case "book":
		fallthrough
	case "manga":
		return c.handleBook(s, i, subcommand)
	case "anime":
		return c.handleAnime(s, i, subcommand)
	default:
		return errors.New("invalid subcommand")
	}
}

func (c *LogCommand) getUserAndTouchGuild(i *discordgo.InteractionCreate) (*users.User, error) {
	var userId string

	if i.GuildID == "" {
		userId = i.User.ID
	} else {
		userId = i.Member.User.ID
	}

	//user, err := c.userRepo.FindOrCreate(context.Background(), userId)
	user, err := c.userState.GetUser(userId)

	if err != nil {
		return nil, err
	}

	if i.GuildID != "" {
		go func() {
			found := false
			for _, guildID := range user.ActiveGuilds {
				if guildID == i.GuildID {
					found = true
					break
				}
			}

			if !found {
				user.ActiveGuilds = append(user.ActiveGuilds, i.GuildID)

				if err = c.userState.UpdateUser(user); err != nil {
					log.Printf("Failed to append active guild: %s->%s\n", userId, i.GuildID)
					log.Printf("Error: %v\n", err)
				}
			}
		}()
	}
	return user, nil
}

func (c *LogCommand) handleAnime(s *discordgo.Session, i *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	user, err := c.getUserAndTouchGuild(i)

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()
	activity.Name = discordutil.GetRequiredStringOption(args, "name")
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeAnime)
	activity.UserID = user.ID

	episodeCount := discordutil.GetRequiredUintOption(args, "episodes")
	episodeDuration := discordutil.GetUintOptionOrDefault(args, "episode-duration", 24)
	duration := episodeDuration * episodeCount
	activity.Duration = time.Duration(duration) * time.Minute
	activity.Meta = map[string]interface{}{"episodes": episodeCount}

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		return err
	}

	activity.Date = date

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		AddField("Episodes Watched", fmt.Sprintf("%d", episodeCount), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess).
		Build()

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})

	return err
}

func (c *LogCommand) handleBook(s *discordgo.Session, i *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	user, err := c.getUserAndTouchGuild(i)

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()
	activity.Name = discordutil.GetRequiredStringOption(args, "name")
	activity.PrimaryType = activities.ActivityImmersionTypeReading
	if subcommand.Name == "book" {
		activity.MediaType = ref.New(activities.ActivityMediaTypeBook)
	} else {
		activity.MediaType = ref.New(activities.ActivityMediaTypeManga)
	}
	activity.UserID = user.ID

	pageCount := discordutil.GetRequiredUintOption(args, "pages")
	duration := discordutil.GetUintOption(args, "duration")

	if pageCount == 0 && duration == nil {
		_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "You must provide either a page count or a duration.",
		})
		return err
	}

	var durationMinutes float64

	if duration != nil {
		durationMinutes = float64(discordutil.GetRequiredUintOption(args, "duration"))
	} else {
		durationMinutes = float64(pageCount) / 2.0
	}

	// because time.Duration casts to uint64, we need to convert to seconds first
	activity.Duration = time.Duration(durationMinutes*60.0) * time.Second

	if pageCount != 0 {
		activity.Meta = map[string]interface{}{"pages": pageCount}
	}

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Invalid date provided.",
		})
		return err
	}

	activity.Date = date

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess)

	if pageCount != 0 {
		embed.AddField("Pages Read", fmt.Sprintf("%d", pageCount), false)
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	})

	return err
}

func (c *LogCommand) handleVisualNovel(s *discordgo.Session, i *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	user, err := c.getUserAndTouchGuild(i)

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()
	activity.Name = discordutil.GetRequiredStringOption(args, "name")
	activity.PrimaryType = activities.ActivityImmersionTypeReading
	activity.MediaType = ref.New(activities.ActivityMediaTypeVisualNovel)
	activity.UserID = user.ID

	charCount := discordutil.GetRequiredUintOption(args, "characters")
	duration := discordutil.GetUintOption(args, "duration")
	readingSpeed := discordutil.GetUintOption(args, "reading-speed")
	readingSpeedHourly := discordutil.GetUintOption(args, "reading-speed-hourly")

	var durationMinutes float64

	if charCount == 0 && duration == nil {
		_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "You must provide either a character count or a duration.",
		})
		return err
	}

	if duration != nil {
		durationMinutes = float64(*duration)
	} else if readingSpeed != nil {
		durationMinutes = float64(charCount) / float64(*readingSpeed)
	} else if readingSpeedHourly != nil {
		durationMinutes = float64(charCount) / (float64(*readingSpeedHourly) / 60.0)
	} else {
		durationMinutes = float64(charCount) / 150.0
	}

	// because time.Duration casts to uint64, we need to convert to seconds first
	activity.Duration = time.Duration(durationMinutes*60.0) * time.Second

	if charCount != 0 {
		activity.Meta = map[string]interface{}{"characters": charCount}
	}

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Invalid date provided.",
		})
		return err
	}

	activity.Date = date

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess)

	if charCount != 0 {
		embed.AddField("Characters Read", fmt.Sprintf("%d", charCount), false)
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	})

	return err
}

func (c *LogCommand) handleVideo(s *discordgo.Session, i *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	user, err := c.getUserAndTouchGuild(i)

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()
	URL := discordutil.GetRequiredStringOption(args, "url")

	u, err := url.Parse(URL)

	if err != nil {
		return err
	}

	video, err := activities.GetVideoInfo(context.Background(), u, false)

	if err != nil {
		return err
	}

	activity.Name = video.Title
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeVideo)
	activity.UserID = user.ID
	activity.Meta = video

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: "Invalid date provided.",
		})
		return err
	}

	activity.Date = date

	if discordutil.GetUintOption(args, "duration") != nil {
		activity.Duration = time.Duration(discordutil.GetRequiredUintOption(args, "duration")) * time.Minute
	} else {
		activity.Duration = video.Duration
	}

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	shareUrl, err := removeTimeParameter(URL)
	if err != nil {
		return err
	}

	durationString := activity.Duration.String()

	if activity.Duration != video.Duration {
		durationString = fmt.Sprintf("%s / %s", activity.Duration.String(), video.Duration.String())
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", video.Title, false).
		AddField("Channel", video.ChannelName, false).
		AddField("Duration Watched", durationString, false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetImage(video.Thumbnail).
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess)

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label: "Video",
				Style: discordgo.LinkButton,
				URL:   shareUrl,
			},
		},
	}

	if video.Platform == "youtube" {
		row.Components = append(row.Components, discordgo.Button{
			Label: "Channel",
			Style: discordgo.LinkButton,
			URL:   fmt.Sprintf("https://www.youtube.com/channel/%s", video.ChannelID),
		})
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
		Components: []discordgo.MessageComponent{
			row,
		},
	})

	return err
}

func (c *LogCommand) handleManual(s *discordgo.Session, i *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	user, err := c.getUserAndTouchGuild(i)

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()
	activity.Name = discordutil.GetRequiredStringOption(args, "name")
	activity.PrimaryType = discordutil.GetRequiredStringOption(args, "type")
	activity.Duration = time.Duration(discordutil.GetRequiredUintOption(args, "duration")) * time.Minute
	activity.MediaType = discordutil.GetStringOption(args, "media-type")
	activity.UserID = user.ID

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid date provided.",
			},
		})
	}

	activity.Date = date

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess).
		Build()

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func removeTimeParameter(urlString string) (string, error) {
	u, err := url.Parse(urlString)

	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Del("t")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func parseDate(enteredDate, timezone *string) (time.Time, error) {
	if enteredDate != nil {
		var parsedDate time.Time
		var err error
		if timezone != nil {
			c := carbon.SetTimezone(*timezone).Parse(*enteredDate)
			err = c.Error
			if err == nil {
				parsedDate = c.ToStdTime()
			}
		} else {
			parsedDate, err = time.Parse("2006-01-02", *enteredDate)
		}
		return parsedDate, err
	} else {
		return time.Now(), nil
	}
}
