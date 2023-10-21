package activities

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/UTD-JLA/botsu/pkg/ytchannel"
	"github.com/kkdai/youtube/v2"
	"github.com/wader/goutubedl"
)

var ytClient = youtube.Client{}

// video is either youtube.com/watch?v=ID or youtube.com/live/ID (for live streams) or youtu.be/ID
var ytVideoLinkRegex = regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtube\.com/live/|youtu\.be/)([a-zA-Z0-9_-]+)`)
var ytHandleRegex = regexp.MustCompile(`(^|\s|youtu.*/)@([a-zA-Z0-9_-]+)($|\s)`)

var channelCache = sync.Map{}

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

	if v.ChannelName == "" {
		v.ChannelName = info.Uploader
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

	var thumbnailURL string

	if len(video.Thumbnails) > 0 {
		thumbnailURL = video.Thumbnails[0].URL
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
		Thumbnail:     thumbnailURL,
	}

	if v.ChannelHandle == "" {
		if cached, ok := channelCache.Load(video.ChannelID); ok {
			v.ChannelHandle = cached.(string)
		} else {
			channel, err := ytchannel.GetYoutubeChannel(ctx, video.ChannelID)

			if err != nil {
				return nil, err
			}

			v.ChannelHandle = channel.Handle
			channelCache.Store(video.ChannelID, channel.Handle)
		}
	}

	highestRes := uint(0)
	highestResThumbnail := ""

	for _, thumbnail := range video.Thumbnails {
		res := thumbnail.Width * thumbnail.Height
		if res > highestRes {
			highestRes = res
			highestResThumbnail = thumbnail.URL
		}

		if res == 0 {
			highestResThumbnail = thumbnail.URL
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
		relatedChannels = append(relatedChannels, "@"+match[2])
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
