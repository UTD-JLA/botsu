package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/aodb"
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
	activityRepo *activities.ActivityRepository
	userRepo     *users.UserRepository
	ytClient     youtube.Client
}

func NewLogCommand(ar *activities.ActivityRepository, ur *users.UserRepository) *LogCommand {
	return &LogCommand{
		activityRepo: ar,
		userRepo:     ur,
		ytClient:     youtube.Client{},
	}
}

func (c *LogCommand) Handle(ctx *bot.InteractionContext) error {
	if ctx.IsAutocomplete() {
		return c.handleAutocomplete(ctx.Session(), ctx.Interaction())
	}

	subcommand := ctx.Data().Options[0]

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

func (c *LogCommand) handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData()
	focuedOption := discordutil.GetFocusedOption(data.Options[0].Options)

	if focuedOption == nil {
		return nil
	}

	if focuedOption.Name != "name" {
		return nil
	}

	input := focuedOption.StringValue()
	results, err := createAutocompleteResult(input)

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

func (c *LogCommand) getUserAndTouchGuild(ctx context.Context, i *discordgo.InteractionCreate) (*users.User, error) {
	var userId string

	if i.GuildID == "" {
		userId = i.User.ID
	} else {
		userId = i.Member.User.ID
	}

	user, err := c.userRepo.FindOrCreate(ctx, userId)

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
				if err = c.userRepo.AppendActiveGuild(context.Background(), userId, i.GuildID); err != nil {
					log.Printf("Failed to append active guild: %s->%s\n", userId, i.GuildID)
					log.Printf("Error: %v\n", err)
				}
			}
		}()
	}
	return user, nil
}

func (c *LogCommand) handleAnime(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	user, err := c.getUserAndTouchGuild(ctx.Context(), ctx.Interaction())

	if err != nil {
		return err
	}

	args := subcommand.Options
	activity := activities.NewActivity()

	episodeCount := discordutil.GetRequiredUintOption(args, "episodes")
	episodeDuration := discordutil.GetUintOptionOrDefault(args, "episode-duration", 24)
	duration := episodeDuration * episodeCount

	nameInput := discordutil.GetRequiredStringOption(args, "name")
	thumbnail := ""
	var namedSources map[string]string

	if isAutocompletedEntry(nameInput) {
		anime, err := resolveAnimeFromAutocomplete(nameInput)
		if err != nil {
			return err
		}

		activity.Name = anime.Title
		thumbnail = anime.Thumbnail
		activity.Meta = map[string]interface{}{
			"episodes_watched": episodeCount,
			"episode_length":   episodeDuration,
			"sources":          anime.Sources,
			"title":            anime.Title,
			"synonyms":         anime.Synonyms,
			"tags":             anime.Tags,
		}
		namedSources = getNamedSources(anime.Sources)
	} else {
		activity.Name = nameInput
		activity.Meta = map[string]interface{}{
			"episodes_watched": episodeCount,
			"episode_length":   episodeDuration,
		}
	}

	activity.Duration = time.Duration(duration) * time.Minute
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeAnime)
	activity.UserID = user.ID

	date, err := parseDate(discordutil.GetStringOption(args, "date"), user.Timezone)

	if err != nil {
		return err
	}

	activity.Date = date

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
		Build()

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

	user, err := c.getUserAndTouchGuild(ctx.Context(), ctx.Interaction())

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
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "You must provide either a page count or a duration.",
		}, false)
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
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "Invalid date provided.",
		}, false)
		return err
	}

	activity.Date = date

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
		SetColor(discordutil.ColorSuccess)

	if pageCount != 0 {
		embed.AddField("Pages Read", fmt.Sprintf("%d", pageCount), false)
	}

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	}, false)

	return err
}

func (c *LogCommand) handleVisualNovel(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	user, err := c.getUserAndTouchGuild(ctx.Context(), ctx.Interaction())

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
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "You must provide either a character count or a duration.",
		}, false)
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
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "Invalid date provided.",
		}, false)
		return err
	}

	activity.Date = date

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
		SetColor(discordutil.ColorSuccess)

	if charCount != 0 {
		embed.AddField("Characters Read", fmt.Sprintf("%d", charCount), false)
	}

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	}, false)

	return err
}

func (c *LogCommand) handleVideo(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	if err := ctx.DeferResponse(); err != nil {
		return err
	}

	user, err := c.getUserAndTouchGuild(ctx.Context(), ctx.Interaction())

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

	video, err := activities.GetVideoInfo(ctx.Context(), u, false)

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
		_, err := ctx.Followup(&discordgo.WebhookParams{
			Content: "Invalid date provided.",
		}, false)
		return err
	}

	activity.Date = date

	if discordutil.GetUintOption(args, "duration") != nil {
		activity.Duration = time.Duration(discordutil.GetRequiredUintOption(args, "duration")) * time.Minute
	} else {
		activity.Duration = video.Duration
	}

	err = c.activityRepo.Create(ctx.Context(), activity)

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

	_, err = ctx.Followup(&discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{embed.Build()},
		Components: []discordgo.MessageComponent{row},
	}, false)

	return err
}

func (c *LogCommand) handleManual(ctx *bot.InteractionContext, subcommand *discordgo.ApplicationCommandInteractionDataOption) error {
	user, err := c.getUserAndTouchGuild(ctx.ResponseContext(), ctx.Interaction())

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
		return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
			Content: "Invalid date provided.",
		})
	}

	activity.Date = date

	err = c.activityRepo.Create(ctx.ResponseContext(), activity)

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

	return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
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

func createAutocompleteResult(input string) (choices []*discordgo.ApplicationCommandOptionChoice, err error) {
	results, err := aodb.Search(input)

	if err != nil {
		return
	}

	choices = make([]*discordgo.ApplicationCommandOptionChoice, 0, 25)

	for i, result := range results {
		if i >= 25 {
			break
		}

		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  truncateLongString(result.Anime.Title, 100),
			Value: fmt.Sprintf("${%d}", result.Index),
		})
	}

	return
}

func isAutocompletedEntry(input string) bool {
	return len(input) > 0 && strings.HasPrefix(input, "${") && strings.HasSuffix(input, "}")
}

func resolveAnimeFromAutocomplete(input string) (*aodb.Anime, error) {
	if !isAutocompletedEntry(input) {
		return nil, errors.New("invalid input")
	}

	indexStr := input[2 : len(input)-1]
	index, err := strconv.Atoi(indexStr)

	if err != nil {
		return nil, err
	}

	result, err := aodb.GetEntry(index)

	if err != nil {
		return nil, err
	}

	return result, nil
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
