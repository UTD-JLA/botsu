package activities

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/wader/goutubedl"
)

var ytClient = youtube.Client{}

// video is either youtube.com/watch?v=ID or youtube.com/live/ID (for live streams) or youtu.be/ID
var ytVideoLinkRegex = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtube\.com/live/|youtu\.be/)([a-zA-Z0-9_-]+)`)
var ytHandleRegex = regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)

func init() {
	youtube.DefaultClient = youtube.WebClient
	goutubedl.Path = "yt-dlp"
}

type VideoInfo struct {
	URL            string        `json:"url"`
	Platform       string        `json:"platform"`
	ID             string        `json:"video_id"`
	Title          string        `json:"video_title"`
	Duration       time.Duration `json:"video_duration"`
	ChannelID      string        `json:"channel_id"`
	ChannelName    string        `json:"channel_name"`
	ChannelHandle  string        `json:"channel_handle"`
	Thumbnail      string        `json:"thumbnail"`
	LinkedChannels []string      `json:"linked_channels"`
	LinkedVideos   []string      `json:"linked_videos"`
}

// Returns a map[string]interface{} meant to be marshalled as JSON
func GetVideoInfo(ctx context.Context, URL *url.URL, forceYtdlp bool) (*VideoInfo, error) {
	isYoutubeLink := URL.Host == "youtu.be" ||
		URL.Host == "youtube.com" ||
		URL.Host == "www.youtube.com" ||
		URL.Host == "m.youtube.com"

	if !forceYtdlp && isYoutubeLink {
		return getInfoFromYoutube(ctx, URL)
	}

	return getGenericVideoInfo(ctx, URL)
}

func getGenericVideoInfo(ctx context.Context, URL *url.URL) (v *VideoInfo, err error) {
	result, err := goutubedl.New(ctx, URL.String(), goutubedl.Options{
		Type: goutubedl.TypeSingle,
	})

	if err != nil {
		return
	}

	info := result.Info

	v = &VideoInfo{
		URL:           URL.String(),
		Platform:      info.Extractor,
		ID:            info.ID,
		Title:         info.Title,
		Duration:      time.Duration(info.Duration) * time.Second,
		ChannelID:     info.ChannelID,
		ChannelName:   info.Channel,
		ChannelHandle: info.UploaderID,
		Thumbnail:     info.Thumbnail,
	}

	return
}

func getInfoFromYoutube(ctx context.Context, URL *url.URL) (v *VideoInfo, err error) {
	var video *youtube.Video

	if strings.HasPrefix(strings.ToLower(URL.Path), "/live/") {
		parts := strings.Split(URL.Path, "/")
		video, err = ytClient.GetVideoContext(ctx, parts[len(parts)-1])
	} else {
		video, err = ytClient.GetVideoContext(ctx, URL.String())
	}

	if err != nil {
		return
	}

	v = &VideoInfo{
		URL:           URL.String(),
		Platform:      "youtube",
		ID:            video.ID,
		Title:         video.Title,
		Duration:      video.Duration,
		ChannelID:     video.ChannelID,
		ChannelName:   video.Author,
		ChannelHandle: video.ChannelHandle,
		Thumbnail:     video.Thumbnails[0].URL,
	}

	highestRes := uint(0)
	highestResThumbnail := ""

	for _, thumnbnail := range video.Thumbnails {
		res := thumnbnail.Width * thumnbnail.Height
		if res > highestRes {
			highestRes = res
			highestResThumbnail = thumnbnail.URL
		}

		if res == 0 {
			highestResThumbnail = thumnbnail.URL
		}
	}

	v.Thumbnail = highestResThumbnail
	v.LinkedChannels = findRelatedYoutubeChannels(video)
	v.LinkedVideos = findRelatedYoutubeVideos(video)

	return
}

func findRelatedYoutubeChannels(video *youtube.Video) []string {
	relatedChannels := make([]string, 0)
	matches := ytHandleRegex.FindAllStringSubmatch(video.Description, -1)
	for _, match := range matches {
		relatedChannels = append(relatedChannels, "@"+match[1])
	}
	return relatedChannels
}

func findRelatedYoutubeVideos(video *youtube.Video) []string {
	relatedVideos := make([]string, 0)
	matches := ytVideoLinkRegex.FindAllStringSubmatch(video.Description, -1)
	for _, match := range matches {
		relatedVideos = append(relatedVideos, match[1])
	}
	return relatedVideos
}
