package ytchannel_test

import (
	"testing"

	"github.com/UTD-JLA/botsu/pkg/ytchannel"
	"github.com/stretchr/testify/assert"
)

func TestGetChannelFromHandle(t *testing.T) {
	channel, err := ytchannel.GetYoutubeChannel("@HakuiKoyori")

	assert.NoError(t, err)
	assert.Equal(t, "UC6eWCld0KwmyHFbAqK3V-Rw", channel.ID)
	assert.Equal(t, "Koyori ch. 博衣こより - holoX -", channel.Name)
	assert.Equal(t, "@HakuiKoyori", channel.Handle)
	assert.NotZero(t, len(channel.Banners))
	assert.NotZero(t, len(channel.Avatars))
}

func TestGetChannelFromID(t *testing.T) {
	channel, err := ytchannel.GetYoutubeChannel("UC6eWCld0KwmyHFbAqK3V-Rw")

	assert.NoError(t, err)
	assert.Equal(t, "UC6eWCld0KwmyHFbAqK3V-Rw", channel.ID)
	assert.Equal(t, "Koyori ch. 博衣こより - holoX -", channel.Name)
	assert.Equal(t, "@HakuiKoyori", channel.Handle)
	assert.NotZero(t, len(channel.Banners))
	assert.NotZero(t, len(channel.Avatars))
}
