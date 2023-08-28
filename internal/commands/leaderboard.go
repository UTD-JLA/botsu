package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/UTD-JLA/botsu/pkg/ref"
	"github.com/bwmarrin/discordgo"
)

var LeaderboardCommandData = &discordgo.ApplicationCommand{
	Name:         "leaderboard",
	Description:  "View the leaderboard",
	DMPermission: ref.New(false),
	Options:      []*discordgo.ApplicationCommandOption{},
}

type LeaderboardCommand struct {
	r *activities.ActivityRepository
}

func NewLeaderboardCommand(r *activities.ActivityRepository) *LeaderboardCommand {
	return &LeaderboardCommand{r: r}
}

func (c *LeaderboardCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if i.GuildID == "" {
		return errors.New("this command can only be used in a guild")
	}

	topMembers, err := c.r.GetTopMembers(context.Background(), i.GuildID, 10, time.Unix(0, 0), time.Now())

	if err != nil {
		return err
	}

	missingMembers := make([]string, 0, len(topMembers))
	foundMembers := make(map[string]*discordgo.Member)

	for _, m := range topMembers {
		guild, err := s.State.Guild(i.GuildID)

		if err != nil {
			return err
		}

		member, err := s.State.Member(guild.ID, m.UserID)

		if err != nil {
			missingMembers = append(missingMembers, m.UserID)
			continue
		}

		foundMembers[m.UserID] = member
	}

	if len(missingMembers) > 0 {
		nonce, err := discordutil.NewNonce()

		fmt.Printf("Requesting chunk %s\n", nonce)
		//len
		fmt.Printf("Requesting chunk %d\n", len(nonce))

		if err != nil {
			return err
		}

		memberChunk := make(chan []*discordgo.Member, 1)

		removeFunc := s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMembersChunk) {
			fmt.Printf("Received chunk %s\n", e.Nonce)

			if e.Nonce != nonce {
				return
			}

			memberChunk <- e.Members
		})

		defer removeFunc()

		err = s.RequestGuildMembersList(i.GuildID, missingMembers, 0, nonce, false)

		if err != nil {
			return err
		}

		select {
		case <-time.After(5 * time.Second):
			return errors.New("timed out waiting for guild members")
		case members := <-memberChunk:
			for _, m := range members {
				foundMembers[m.User.ID] = m
			}
		}
	}

	embed := discordutil.NewEmbedBuilder().
		SetTitle("Leaderboard").
		SetColor(discordutil.ColorPrimary)

	for _, m := range topMembers {
		member, ok := foundMembers[m.UserID]
		displayName := m.UserID
		if ok && member.Nick != "" {
			displayName = member.Nick
		} else if ok && member.User != nil {
			displayName = member.User.Username
		}

		embed.AddField(displayName, m.TotalDuration.String(), false)
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	})

	return err
}
