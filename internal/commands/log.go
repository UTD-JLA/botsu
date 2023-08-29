package commands

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/UTD-JLA/botsu/pkg/ref"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-module/carbon/v2"
	"github.com/kkdai/youtube/v2"
)

func init() {
	youtube.DefaultClient = youtube.WebClient
}

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
	default:
		return nil
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
				// if err = c.userRepo.AppendActiveGuild(context.Background(), user.ID, i.GuildID); err != nil {
				// 	log.Printf("Failed to append active guild: %s->%s\n", userId, i.GuildID)
				// }

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
	activity.Meta["characters"] = discordutil.GetRequiredUintOption(args, "characters")

	duration := discordutil.GetUintOption(args, "duration")

	if duration != nil {
		activity.Duration = time.Duration(*duration) * time.Minute
	} else {
		readingSpeed := float64(discordutil.GetUintOptionOrDefault(args, "reading-speed", 150))
		readingSpeedHourly := discordutil.GetUintOptionOrDefault(args, "reading-speed-hourly", 0)

		if readingSpeedHourly > 0 {
			readingSpeed = float64(readingSpeedHourly) / 60.0
		}

		mins := float64(activity.Meta["characters"].(uint64)) / readingSpeed
		activity.Duration = time.Duration(mins*60) * time.Second

	}

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
		AddField("Characters Read", fmt.Sprintf("%d", activity.Meta["characters"]), false).
		AddField("Duration", activity.Duration.String(), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess).
		Build()

	// return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
	// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
	// 	Data: &discordgo.InteractionResponseData{
	// 		Embeds: []*discordgo.MessageEmbed{embed},
	// 	},
	// })

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
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
	url := discordutil.GetRequiredStringOption(args, "url")

	video, err := c.ytClient.GetVideo(url)

	if err != nil {
		return err
	}

	activity.Name = video.Title
	activity.PrimaryType = activities.ActivityImmersionTypeListening
	activity.MediaType = ref.New(activities.ActivityMediaTypeVideo)
	activity.UserID = user.ID
	activity.Meta["url"] = url
	activity.Meta["platform"] = "youtube"
	activity.Meta["video_id"] = video.ID
	activity.Meta["video_duration"] = video.Duration
	activity.Meta["channel_id"] = video.ChannelID
	activity.Meta["channel_name"] = video.Author
	activity.Meta["channel_handle"] = video.ChannelHandle
	activity.Meta["video_thumbnails"] = make([]struct {
		Url    string `json:"url"`
		Width  uint   `json:"width"`
		Height uint   `json:"height"`
	}, len(video.Thumbnails))
	activity.Meta["related_channels"] = findRelatedChannels(video)
	activity.Meta["related_videos"] = findRelatedVideos(video)
	activity.Meta["video_publish_date"] = video.PublishDate

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

	if discordutil.GetUintOption(args, "duration") != nil {
		activity.Duration = time.Duration(discordutil.GetRequiredUintOption(args, "duration")) * time.Minute
	} else {
		activity.Duration = video.Duration
	}

	thumbnailUrl := ""
	largestResolution := uint(0)
	largestThumbnail := ""

	for i, thumbnail := range video.Thumbnails {
		if thumbnail.Width > largestResolution {
			largestResolution = thumbnail.Width
			largestThumbnail = thumbnail.URL
		}

		if strings.Contains(thumbnail.URL, "mqdefault.webp") {
			thumbnailUrl = thumbnail.URL
		}

		activity.Meta["video_thumbnails"].([]struct {
			Url    string `json:"url"`
			Width  uint   `json:"width"`
			Height uint   `json:"height"`
		})[i] = struct {
			Url    string `json:"url"`
			Width  uint   `json:"width"`
			Height uint   `json:"height"`
		}{
			Url:    thumbnail.URL,
			Width:  thumbnail.Width,
			Height: thumbnail.Height,
		}
	}

	err = c.activityRepo.Create(context.Background(), activity)

	if err != nil {
		return err
	}

	if thumbnailUrl == "" {
		thumbnailUrl = largestThumbnail
	}

	shareUrl, err := removeTimeParameter(url)

	if err != nil {
		return err
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Activity logged!").
		AddField("Channel", video.Author, false).
		AddField("Title", video.Title, false).
		AddField("Duration Watched", fmt.Sprintf("%s / %s", activity.Duration.String(), video.Duration.String()), false).
		SetFooter(fmt.Sprintf("ID: %d", activity.ID), "").
		SetImage(thumbnailUrl).
		SetTimestamp(activity.Date).
		SetColor(discordutil.ColorSuccess).
		Build()

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label: "Video",
				Style: discordgo.LinkButton,
				URL:   shareUrl,
			},
			discordgo.Button{
				Label: "Channel",
				Style: discordgo.LinkButton,
				URL:   fmt.Sprintf("https://www.youtube.com/channel/%s", video.ChannelID),
			},
		},
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
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

func findRelatedChannels(video *youtube.Video) []string {
	channelIdRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	relatedChannels := make([]string, 0)
	matches := channelIdRegex.FindAllStringSubmatch(video.Description, -1)
	for _, match := range matches {
		relatedChannels = append(relatedChannels, "@"+match[1])
	}
	return relatedChannels
}

func findRelatedVideos(video *youtube.Video) []string {
	// video is either youtube.com/watch?v=ID or youtube.com/live/ID (for live streams) or youtu.be/ID
	videoIdRegex := regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtube\.com/live/|youtu\.be/)([a-zA-Z0-9_-]+)`)
	relatedVideos := make([]string, 0)
	matches := videoIdRegex.FindAllStringSubmatch(video.Description, -1)
	for _, match := range matches {
		relatedVideos = append(relatedVideos, match[1])
	}
	return relatedVideos
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
