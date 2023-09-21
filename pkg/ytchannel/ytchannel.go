package ytchannel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var client = http.DefaultClient
var ytInitialDataRegex = regexp.MustCompile(`var ytInitialData\s*=\s*(\{.+?\});`)

var ErrInvalidChannelIdentifier = errors.New("invalid channel identifier string, you should either provide a handle (starting with '@') or an ID (starting with 'UC')")
var ErrYtInitialDataNotFound = errors.New("could not find ytInitialData")

type navigationEndpoint struct {
	BrowseEndpoint struct {
		CanonicalBaseUrl string `json:"canonicalBaseUrl"`
	}
}

type thumbnailsWrapper struct {
	Thumbnails []*Thumbail `json:"thumbnails"`
}

type headerRenderer struct {
	Title              string             `json:"title"`
	ChannelID          string             `json:"channelId"`
	Avatars            thumbnailsWrapper  `json:"avatar"`
	Banners            thumbnailsWrapper  `json:"banner"`
	NavigationEndpoint navigationEndpoint `json:"navigationEndpoint"`
}

type ytInitialData struct {
	Header struct {
		HeaderRenderer headerRenderer `json:"c4TabbedHeaderRenderer"`
	} `json:"header"`
}

type Thumbail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Channel struct {
	ID      string
	Name    string
	Handle  string
	Avatars []*Thumbail
	Banners []*Thumbail
}

func GetYoutubeChannel(handle string) (ch *Channel, err error) {
	var profileURL string

	if strings.HasPrefix(handle, "@") {
		profileURL = fmt.Sprintf("https://youtube.com/%s", handle)
	} else if strings.HasPrefix(handle, "UC") {
		profileURL = fmt.Sprintf("https://youtube.com/channel/%s", handle)
	} else {
		err = ErrInvalidChannelIdentifier
		return
	}

	resp, err := client.Get(profileURL)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return
	}

	data := ytInitialDataRegex.FindSubmatch(body)

	var initialData ytInitialData

	if len(data) < 2 {
		err = ErrYtInitialDataNotFound
		return
	}

	if err = json.Unmarshal(data[1], &initialData); err != nil {
		return
	}

	if len(initialData.Header.HeaderRenderer.NavigationEndpoint.BrowseEndpoint.CanonicalBaseUrl) > 1 {
		handle = initialData.Header.HeaderRenderer.NavigationEndpoint.BrowseEndpoint.CanonicalBaseUrl[1:]
	}

	ch = &Channel{
		Handle:  handle,
		Name:    initialData.Header.HeaderRenderer.Title,
		ID:      initialData.Header.HeaderRenderer.ChannelID,
		Avatars: initialData.Header.HeaderRenderer.Avatars.Thumbnails,
		Banners: initialData.Header.HeaderRenderer.Banners.Thumbnails,
	}

	return
}
