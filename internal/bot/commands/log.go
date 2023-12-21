package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/guilds"
	"github.com/UTD-JLA/botsu/internal/mediadata"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/UTD-JLA/botsu/pkg/ref"
	"github.com/bwmarrin/discordgo"
	"github.com/golang-module/carbon/v2"
	"github.com/jackc/pgx/v5"
	"github.com/kkdai/youtube/v2"
)

var (
	errInvalidDateInput              = errors.New("invalid date")
	errInvalidMediaAutocompleteInput = errors.New("invalid media autocomplete input")
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
	{
		Name:        "complex-duration",
		Type:        discordgo.ApplicationCommandOptionString,
		Description: "Duration spent on the activity",
		Required:    false,
		Options:     []*discordgo.ApplicationCommandOption{},
	},
}

var vnCommandOptions = []*discordgo.ApplicationCommandOption{
	{
		Name:         "name",
		Type:         discordgo.ApplicationCommandOptionString,
		Description:  "Title/name of the book read",
		Required:     true,
		Autocomplete: true,
		Options:      []*discordgo.ApplicationCommandOption{},
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
		Name:         "name",
		Type:         discordgo.ApplicationCommandOptionString,
		Description:  "Title/name of the anime watched",
		Required:     true,
		Options:      []*discordgo.ApplicationCommandOption{},
		Autocomplete: true,
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
	activityRepo  *activities.ActivityRepository
	userRepo      *users.UserRepository
	guildRepo     *guilds.GuildRepository
	mediaSearcher *mediadata.MediaSearcher
	ytClient      youtube.Client
}

func NewLogCommand(
	ar *activities.ActivityRepository,
	ur *users.UserRepository,
	gr *guilds.GuildRepository,
	ms *mediadata.MediaSearcher,
) *LogCommand {
	return &LogCommand{
		activityRepo:  ar,
		userRepo:      ur,
		mediaSearcher: ms,
		guildRepo:     gr,
		ytClient:      youtube.Client{},
	}
}

func (c *LogCommand) Handle(ctx *bot.InteractionContext) error {
	if ctx.IsAutocomplete() {
		return c.handleAutocomplete(ctx.ResponseContext(), ctx.Session(), ctx.Interaction())
	}

	if len(ctx.Options()) == 0 {
		return bot.ErrInvalidOptions
	}

	subcommand := ctx.Options()[0]

	switch subcommand.Name {
	case "manual":
		return c.handleManual(ctx, subcommand)
	case "video":
		return c.handleVideo(ctx, subcommand)
	case "vn":
		return c.handleVisualNovel(ctx, subcommand)
	case "book":
		fallthrough
	case "manga":
		return c.handleBook(ctx, subcommand)
	case "anime":
		return c.handleAnime(ctx, subcommand)
	default:
		return errors.New("invalid subcommand")
	}
}

func (c *LogCommand) handleAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()

	if len(data.Options) == 0 {
		return bot.ErrInvalidOptions
	}

	subcommand := data.Options[0].Name
	focusedOption := discordutil.GetFocusedOption(data.Options[0].Options)

	if focusedOption == nil {
		return nil
	}

	if focusedOption.Name != "name" {
		return nil
	}

	var mediaType string

	if subcommand == "anime" {
		mediaType = activities.ActivityMediaTypeAnime
	} else if subcommand == "vn" {
		mediaType = activities.ActivityMediaTypeVisualNovel
	}

	input := focusedOption.StringValue()
	results, err := c.createAutocompleteResult(ctx, mediaType, input)

	if err != nil {
		return err
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: results,
		},
	})
}

func (c *LogCommand) handleAnime(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	args := subcommand.Options
	activity := activities.NewActivity()

	if ctx.Interaction().GuildID != "" {
		activity.GuildID = &ctx.Interaction().GuildID
	}

	episodeCount, err := discordutil.GetRequiredUintOption(args, "episodes")
	if err != nil {
		return err
	}
	episodeDuration := discordutil.GetUintOptionOrDefault(args, "episode-duration", 24)
	duration := episodeDuration * episodeCount

	nameInput, err := discordutil.GetRequiredStringOption(args, "name")
	if err != nil {
		return err
	}
	thumbnail := ""
	var namedSources map[string]string

	if isAutocompletedEntry(nameInput) {
		anime, titleField, err := c.resolveAnimeFromAutocomplete(nameInput)
		if err != nil {
			return err
		}

		activity.Name = anime.PrimaryTitle

		if titleField == "jp" {
			activity.Name = anime.JapaneseOfficialTitle
		} else if titleField == "en" {
			activity.Name = anime.EnglishOfficialTitle
		} else if titleField == "x-jat" {
			activity.Name = anime.RomajiOfficialTitle
		}

		thumbnail = anime.Thumbnail
		activity.SetMeta("anidb_id", anime.ID)
		activity.SetMeta("thumbnail", anime.Thumbnail)
		activity.SetMeta("sources", anime.Sources)
		activity.SetMeta("title", anime.PrimaryTitle)
		activity.SetMeta("tags", anime.Tags)
		namedSources = getNamedSources(anime.Sources)
	} else {
		activity.Name = nameInput
	}

	activity.SetMeta("episodes", episodeCount)

	activity.Duration = time.Duration(duration) * time.Minute
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeAnime)
	activity.UserID = userID

	if activity.Date, err = c.parseDateOption(ctx, args); err != nil {
		if errors.Is(err, errInvalidDateInput) {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: "Invalid date provided.",
			}, false)
		}

		return err
	}

	err = c.activityRepo.Create(ctx.Context(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		AddField("Episodes Watched", fmt.Sprintf("%d", episodeCount), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetThumbnail(thumbnail).
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess).
		MessageEmbed

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{},
	}

	for name, url := range namedSources {
		row.Components = append(row.Components, discordgo.Button{
			Label:    name,
			Style:    discordgo.LinkButton,
			Disabled: false,
			URL:      url,
		})
	}

	params := discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}

	if len(row.Components) > 0 {
		params.Components = []discordgo.MessageComponent{row}
	}

	_, err = ctx.Followup(&params, false)

	return err
}

func (c *LogCommand) handleBook(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	args := subcommand.Options
	activity := activities.NewActivity()

	if ctx.Interaction().GuildID != "" {
		activity.GuildID = &ctx.Interaction().GuildID
	}

	var err error
	if activity.Name, err = discordutil.GetRequiredStringOption(args, "name"); err != nil {
		return err
	}
	activity.PrimaryType = activities.ActivityImmersionTypeReading
	if subcommand.Name == "book" {
		activity.MediaType = ref.New(activities.ActivityMediaTypeBook)
	} else {
		activity.MediaType = ref.New(activities.ActivityMediaTypeManga)
	}
	activity.UserID = userID

	pageCount, err := discordutil.GetRequiredUintOption(args, "pages")
	if err != nil {
		return err
	}
	duration := discordutil.GetUintOption(args, "duration")

	if pageCount == 0 && duration == nil {
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "You must provide either a page count or a duration.",
		}, false)
		return err
	}

	var durationMinutes float64

	if duration != nil && pageCount != 0 {
		// if both duration and page count is provided
		durationMinutes = float64(*duration)
		activity.SetMeta("pages", pageCount)
		activity.SetMeta("speed", float64(pageCount)/(durationMinutes))
	} else if pageCount != 0 {
		// if only page count is provided
		durationMinutes = float64(pageCount) / 2.0
		activity.SetMeta("pages", pageCount)
	} else {
		// if only duration is provided
		durationMinutes = float64(*duration)
	}

	// because time.Duration casts to uint64, we need to convert to seconds first
	activity.Duration = time.Duration(durationMinutes*60.0) * time.Second

	if activity.Date, err = c.parseDateOption(ctx, args); err != nil {
		if errors.Is(err, errInvalidDateInput) {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: "Invalid date provided.",
			}, false)
		}

		return err
	}

	if err = c.activityRepo.Create(ctx.Context(), activity); err != nil {
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

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.MessageEmbed},
	}, false)

	return err
}

func (c *LogCommand) handleVisualNovel(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	args := subcommand.Options
	activity := activities.NewActivity()

	if ctx.Interaction().GuildID != "" {
		activity.GuildID = &ctx.Interaction().GuildID
	}

	var err error
	if activity.Name, err = discordutil.GetRequiredStringOption(args, "name"); err != nil {
		return err
	}
	activity.PrimaryType = activities.ActivityImmersionTypeReading
	activity.MediaType = ref.New(activities.ActivityMediaTypeVisualNovel)
	activity.UserID = userID

	charCount, err := discordutil.GetRequiredUintOption(args, "characters")
	if err != nil {
		return err
	}
	duration := discordutil.GetUintOption(args, "duration")
	readingSpeed := discordutil.GetUintOption(args, "reading-speed")
	readingSpeedHourly := discordutil.GetUintOption(args, "reading-speed-hourly")

	thumbnail := ""

	if isAutocompletedEntry(activity.Name) {
		v, titleField, err := c.resolveVNFromAutocomplete(activity.Name)
		if err != nil {
			return err
		}

		if titleField == "jp" {
			activity.Name = v.JapaneseTitle
		} else if titleField == "en" {
			activity.Name = v.EnglishTitle
		} else if titleField == "romaji" {
			activity.Name = v.RomajiTitle
		}

		thumbnail = v.ImageURL()
		activity.SetMeta("vndb_id", v.ID)
		activity.SetMeta("thumbnail", v.ImageURL())

	}

	if charCount != 0 {
		activity.SetMeta("characters", charCount)
	}

	var durationMinutes float64

	if charCount == 0 && duration == nil {
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "You must provide either a character count or a duration.",
		}, false)
		return err
	}

	speedIsKnown := true

	if duration != nil {
		durationMinutes = float64(*duration)
	} else if readingSpeed != nil {
		durationMinutes = float64(charCount) / float64(*readingSpeed)
	} else if readingSpeedHourly != nil {
		durationMinutes = float64(charCount) / (float64(*readingSpeedHourly) / 60.0)
	} else {
		durationMinutes = float64(charCount) / 150.0
		speedIsKnown = false
	}

	if charCount != 0 && speedIsKnown {
		activity.SetMeta("speed", float64(charCount)/(durationMinutes))
	}

	// because time.Duration casts to uint64, we need to convert to seconds first
	activity.Duration = time.Duration(durationMinutes*60.0) * time.Second

	if activity.Date, err = c.parseDateOption(ctx, args); err != nil {
		if errors.Is(err, errInvalidDateInput) {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: "Invalid date provided.",
			}, false)
		}

		return err
	}

	err = c.activityRepo.Create(ctx.Context(), activity)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Title", activity.Name, false).
		AddField("Duration", activity.Duration.String(), false).
		SetThumbnail(thumbnail).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess)

	if charCount != 0 {
		embed.AddField("Characters Read", fmt.Sprintf("%d", charCount), false)
	}

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.MessageEmbed},
	}, false)

	return err
}

func (c *LogCommand) handleVideo(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	args := subcommand.Options
	activity := activities.NewActivity()

	if ctx.Interaction().GuildID != "" {
		activity.GuildID = &ctx.Interaction().GuildID
	}

	URL, err := discordutil.GetRequiredStringOption(args, "url")

	if err != nil {
		return err
	}

	u, err := url.Parse(URL)

	if err != nil {
		return err
	}

	video, err := activities.GetVideoInfo(ctx.Context(), u, false)

	if err != nil {
		return err
	}

	activity.Name = video.Title
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeVideo)
	activity.UserID = userID
	activity.Meta = video

	if activity.Date, err = c.parseDateOption(ctx, args); err != nil {
		if errors.Is(err, errInvalidDateInput) {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: "Invalid date provided.",
			}, false)
		}

		return err
	}

	if durationMinutes := discordutil.GetUintOption(args, "duration"); durationMinutes != nil {
		activity.Duration = time.Duration(*durationMinutes) * time.Minute
	} else if complexDuration := discordutil.GetStringOption(args, "complex-duration"); complexDuration != nil {
		var lowerDuration time.Duration
		var tDuration time.Duration

		if tSeconds, err := strconv.Atoi(u.Query().Get("t")); err == nil {
			tDuration = time.Second * time.Duration(tSeconds)
		}

		lowerDuration, err = c.activityRepo.GetTotalWatchTimeOfVideoByUserID(ctx.Context(), userID, video.Platform, video.ID)

		if err != nil {
			return err
		}

		// TODO: lazy loading variables to prevent unneeded db queries
		vars := map[string]time.Duration{
			"t": tDuration,
			"_": lowerDuration,
		}

		activity.Duration, err = parseDurationComplex(*complexDuration, video.Duration, vars)

		if err != nil {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: fmt.Sprintf("Invalid duration provided: %s", err.Error()),
			}, false)
			return err
		}

		if activity.Duration < 0 {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: fmt.Sprintf("Expected positive duration, got %s", activity.Duration.String()),
			}, false)
			return err
		}
	} else {
		activity.Duration = video.Duration
	}

	err = c.activityRepo.Create(ctx.Context(), activity)

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
				URL:   URL,
			},
		},
	}

	if video.Platform == "youtube" {
		// make new shorturl to get rid of parameters (such as t and sid)
		shortURL := fmt.Sprintf("https://youtu.be/%s", video.ID)

		row.Components = []discordgo.MessageComponent{
			discordgo.Button{
				Label: "Video",
				Style: discordgo.LinkButton,
				URL:   shortURL,
			},
			discordgo.Button{
				Label: "Channel",
				Style: discordgo.LinkButton,
				URL:   fmt.Sprintf("https://www.youtube.com/channel/%s", video.ChannelID),
			},
		}
	}

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed.MessageEmbed},
		Components: []discordgo.MessageComponent{row},
	}, false)

	return err
}

func (c *LogCommand) handleManual(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID

	args := subcommand.Options
	activity := activities.NewActivity()

	if ctx.Interaction().GuildID != "" {
		activity.GuildID = &ctx.Interaction().GuildID
	}

	var err error
	if activity.Name, err = discordutil.GetRequiredStringOption(args, "name"); err != nil {
		return err
	}
	if activity.PrimaryType, err = discordutil.GetRequiredStringOption(args, "type"); err != nil {
		return err
	}
	durationOption, err := discordutil.GetRequiredUintOption(args, "duration")
	if err != nil {
		return err
	}
	activity.Duration = time.Duration(durationOption) * time.Minute
	activity.MediaType = discordutil.GetStringOption(args, "media-type")
	activity.UserID = userID

	if activity.Date, err = c.parseDateOption(ctx, args); err != nil {
		if errors.Is(err, errInvalidDateInput) {
			_, err = ctx.Followup(&discordgo.WebhookParams{
				Content: "Invalid date provided.",
			}, false)
		}

		return err
	}

	err = c.activityRepo.Create(ctx.Context(), activity)

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
		MessageEmbed

	return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
}

func (c *LogCommand) createAutocompleteResult(ctx context.Context, mediaType, input string) (choices []*discordgo.ApplicationCommandOptionChoice, err error) {
	choices = make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

	if mediaType == activities.ActivityMediaTypeAnime {
		results, err := c.mediaSearcher.SearchAnime(ctx, input, 25)

		if err != nil {
			return nil, err
		}

		for _, result := range results {
			var title string
			var fieldID string

			switch result.Field {
			case mediadata.AnimeSearchFieldEnglishOfficialTitle:
				title = result.Value.EnglishOfficialTitle
				fieldID = "en"
			case mediadata.AnimeSearchFieldJapaneseOfficialTitle:
				title = result.Value.JapaneseOfficialTitle
				fieldID = "jp"
			case mediadata.AnimeSearchFieldRomajiOfficialTitle:
				title = result.Value.RomajiOfficialTitle
				fieldID = "x-jat"
			default:
				title = result.Value.PrimaryTitle
				fieldID = "primary"
			}

			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  truncateLongString(title, 100),
				Value: fmt.Sprintf("${%s:%s}", result.Value.ID, fieldID),
			})
		}
	} else if mediaType == activities.ActivityMediaTypeVisualNovel {
		results, err := c.mediaSearcher.SearchVisualNovel(ctx, input, 25)

		if err != nil {
			return nil, err
		}

		for _, result := range results {
			var title string
			var fieldID string

			switch result.Field {
			case mediadata.VNSearchFieldJapaneseTitle:
				title = result.Value.JapaneseTitle
				fieldID = "jp"
			case mediadata.VNSearchFieldEnglishTitle:
				title = result.Value.EnglishTitle
				fieldID = "en"
			case mediadata.VNSearchFieldRomajiTitle:
				title = result.Value.RomajiTitle
				fieldID = "romaji"
			}

			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  truncateLongString(title, 100),
				Value: fmt.Sprintf("${%s:%s}", result.Value.ID, fieldID),
			})
		}
	}

	return
}

func (c *LogCommand) resolveAnimeFromAutocomplete(input string) (*mediadata.Anime, string, error) {
	if !isAutocompletedEntry(input) {
		return nil, "", errInvalidMediaAutocompleteInput
	}

	idAndField := input[2 : len(input)-1]

	parts := strings.Split(idAndField, ":")
	if len(parts) != 2 {
		return nil, "", errInvalidMediaAutocompleteInput
	}

	id, field := parts[0], parts[1]

	anime, err := c.mediaSearcher.ReadAnime(context.TODO(), id)

	if err != nil {
		return nil, "", err
	}

	return anime, field, nil
}

func (c *LogCommand) resolveVNFromAutocomplete(input string) (*mediadata.VisualNovel, string, error) {
	if !isAutocompletedEntry(input) {
		return nil, "", errInvalidMediaAutocompleteInput
	}

	idAndField := input[2 : len(input)-1]

	parts := strings.Split(idAndField, ":")

	if len(parts) != 2 {
		return nil, "", errInvalidMediaAutocompleteInput
	}

	id, field := parts[0], parts[1]

	vn, err := c.mediaSearcher.ReadVisualNovel(context.TODO(), id)

	if err != nil {
		return nil, "", err
	}

	return vn, field, nil
}

func (c *LogCommand) parseDateOption(
	ctx *bot.InteractionContext,
	args []*discordgo.ApplicationCommandInteractionDataOption,
) (date time.Time, err error) {
	date = time.Now()
	enteredDate := discordutil.GetStringOption(args, "date")

	if enteredDate == nil {
		return
	}

	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID
	guildID := ctx.Interaction().GuildID

	user, err := c.userRepo.FindOrCreate(ctx.Context(), userID)

	if err != nil {
		return
	}

	timezone := carbon.UTC

	if user != nil && user.Timezone != nil {
		timezone = *user.Timezone
	} else if guildID != "" {
		var guild *guilds.Guild
		guild, err = c.guildRepo.FindByID(ctx.Context(), guildID)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return
		}

		if guild != nil && guild.Timezone != nil {
			timezone = *guild.Timezone
		}
	}

	cb := carbon.SetTimezone(timezone).Parse(*enteredDate)

	if err = cb.Error; err != nil {
		err = fmt.Errorf("%w: %s", errInvalidDateInput, err.Error())
		return
	}

	date = cb.ToStdTime()

	return
}

// Returns difference between two durations
// Should be in one of the following formats:
// duration1:duration2
// (:?)duration2 			(duration1 is inferred to be 0)
// duration1: 				(duration2 is inferred to be upper)
// or _ as duration1 for lower (ex "_:" is lower:upper)
func parseDurationComplex(str string, upper time.Duration, vars map[string]time.Duration) (d time.Duration, err error) {
	parts := strings.Split(str, ":")

	durationOrVar := func(durationString string) (d time.Duration, err error) {
		if len(durationString) < 1 {
			err = errors.New("expected duration string to not be empty")
			return
		}

		switch durationString[0] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-', '+', '.':
			d, err = time.ParseDuration(durationString)
		default:
			var ok bool
			if d, ok = vars[durationString]; !ok {
				err = fmt.Errorf("unknown duration variable: %s", durationString)
			}
		}

		return
	}

	// duration2
	if len(parts) == 1 {
		d, err = durationOrVar(parts[0])
		return
	}

	if len(parts) != 2 {
		err = errors.New("invalid duration")
		return
	}

	// :
	if parts[0] == "" && parts[1] == "" {
		d = upper
		return
	}

	// :duration2
	if parts[0] == "" {
		d, err = durationOrVar(parts[1])
		if d < 0 {
			d = upper - d.Abs()
		}
		return
	}

	// duration1:
	if parts[1] == "" {
		d, err = durationOrVar(parts[0])
		d = upper - d
		return
	}

	// duration1:duration2
	d1, err := durationOrVar(parts[0])
	if err != nil {
		return
	}

	d2, err := durationOrVar(parts[1])
	if err != nil {
		return
	}

	if d2 < 0 {
		d2 = upper - d.Abs()
	}

	d = d2 - d1
	return
}

func getNamedSources(sources []string) map[string]string {
	result := make(map[string]string)
	for _, source := range sources {
		if strings.HasPrefix(source, "https://myanimelist.net/") {
			result["MAL"] = source
		} else if strings.HasPrefix(source, "https://anilist.co/") {
			result["AniList"] = source
		} else if strings.HasPrefix(source, "https://kitsu.io/") {
			result["Kitsu"] = source
		} else if strings.HasPrefix(source, "https://anidb.net/") {
			result["AniDB"] = source
		}
	}

	// remove kitsu and anidb if they are not the only sources
	if len(result) > 2 {
		delete(result, "Kitsu")
		delete(result, "AniDB")
	}

	return result
}

func truncateLongString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func isAutocompletedEntry(input string) bool {
	return len(input) > 0 && strings.HasPrefix(input, "${") && strings.HasSuffix(input, "}")
}
