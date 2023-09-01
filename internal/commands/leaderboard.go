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
	"github.com/golang-module/carbon/v2"
)

var LeaderboardCommandData = &discordgo.ApplicationCommand{
	Name:         "leaderboard",
	Description:  "View the leaderboard",
	DMPermission: ref.New(false),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "day",
			Description: "View the leaderboard for the current day",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "week",
			Description: "View the leaderboard for the current week",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "month",
			Description: "View the leaderboard for the current month",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "year",
			Description: "View the leaderboard for the current year",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "all",
			Description: "View the leaderboard for all time",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "custom",
			Description: "View the leaderboard over a custom time period",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "start",
					Description: "The start date",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "end",
					Description: "The end date",
					Required:    true,
				},
			},
		},
	},
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

	var start, end time.Time
	now := carbon.Now(carbon.UTC)

	switch i.ApplicationCommandData().Options[0].Name {
	case "day":
		start = now.StartOfDay().ToStdTime()
		end = now.EndOfDay().ToStdTime()
	case "week":
		start = now.StartOfWeek().ToStdTime()
		end = now.EndOfWeek().ToStdTime()
	case "month":
		start = now.StartOfMonth().ToStdTime()
		end = now.EndOfMonth().ToStdTime()
	case "year":
		start = now.StartOfYear().ToStdTime()
		end = now.EndOfYear().ToStdTime()
	case "all":
		start = time.Unix(0, 0)
		end = time.Now()
	case "custom":
		options := i.ApplicationCommandData().Options[0].Options
		user, err := c.u.FindByID(context.Background(), i.Member.User.ID)

		if err != nil {
			return err
		}

		timezone := carbon.UTC

		if user != nil && user.Timezone != nil {
			timezone = *user.Timezone
		}

		startString := discordutil.GetRequiredStringOption(options, "start")
		endString := discordutil.GetRequiredStringOption(options, "end")
		carbonStart := carbon.SetTimezone(timezone).Parse(startString)
		carbonEnd := carbon.SetTimezone(timezone).Parse(endString)

		validStart := carbonStart.IsValid()
		validEnd := carbonEnd.IsValid()
		errorMsg := ""

		if !validStart && !validEnd {
			errorMsg = "Invalid start and end date."
		} else if !validStart {
			errorMsg = "Invalid start date."
		} else if !validEnd {
			errorMsg = "Invalid end date."
		}

		if errorMsg != "" {
			_, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: errorMsg,
			})

			return err
		}

		start = carbonStart.ToStdTime()
		end = carbonEnd.ToStdTime()
	}

	// Note: Do not go over 100 members as Discord will not allow fetching 100+ in a single chunk
	topMembers, err := c.r.GetTopMembers(context.Background(), i.GuildID, 10, start, end)

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

	description := fmt.Sprintf("Starting <t:%d:R>, resetting <t:%d:R>.", start.Unix(), end.Unix())

	embed := discordutil.NewEmbedBuilder().
		SetDescription(description).
		SetTitle("Leaderboard").
		SetColor(discordutil.ColorPrimary).
		SetTimestamp(time.Now())

	deadMembers := make([]string, 0, len(topMembers))

	// if isRollingLeaderboard {
	// 	embed.SetFooter("Next reset", "")
	// 	embed.SetTimestamp(end)
	// }

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
