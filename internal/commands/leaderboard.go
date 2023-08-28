package commands

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/users"
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
	u *users.UserRepository
}

func NewLeaderboardCommand(r *activities.ActivityRepository, u *users.UserRepository) *LeaderboardCommand {
	return &LeaderboardCommand{r: r, u: u}
}

func (c *LeaderboardCommand) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	if err != nil {
		return err
	}

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

		if err != nil {
			return err
		}

		memberChunk := make(chan []*discordgo.Member, 1)

		removeHandler := s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMembersChunk) {
			if e.Nonce == nonce {
				memberChunk <- e.Members
			}
		})

		defer removeHandler()

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

	deadMembers := make([]string, 0, len(topMembers))

	for x, m := range topMembers {
		member, ok := foundMembers[m.UserID]
		displayName := m.UserID
		if ok && member.Nick != "" {
			displayName = member.Nick
		} else if ok && member.User != nil {
			displayName = member.User.Username
		} else if !ok {
			deadMembers = append(deadMembers, m.UserID)
			usr, err := s.User(m.UserID)
			if err != nil {
				log.Printf("Error getting user %s: %v\n", m.UserID, err)
				continue
			}

			displayName = usr.Username
		}

		embed.AddField(fmt.Sprintf("%d. %s", x+1, displayName), m.TotalDuration.String(), false)
	}

	_, err = s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed.Build()},
	})

	go func() {
		for _, m := range deadMembers {
			err := c.u.RemoveActiveGuild(context.Background(), m, i.GuildID)
			if err != nil {
				log.Printf("Error removing dead member %s: %v\n", m, err)
			}
		}
	}()

	return err
}
