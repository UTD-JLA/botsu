package discordutil

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GetFocusedOption(options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Focused {
			return option
		}
	}

	return nil
}

func GetOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Name == key {
			return option
		}
	}

	return nil
}

func GetRequiredOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *discordgo.ApplicationCommandInteractionDataOption {
	option := GetOption(options, key)
	if option == nil {
		panic(fmt.Sprintf("Required option %s not found", key))
	}

	return option
}

func GetStringOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *string {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	str := option.StringValue()
	return &str
}

func GetStringOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key, defaultValue string) string {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.StringValue()
}

func GetRequiredStringOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) string {
	option := GetRequiredOption(options, key)
	return option.StringValue()
}

func GetIntOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *int64 {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	i := option.IntValue()
	return &i
}

func GetIntOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue int64) int64 {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.IntValue()
}

func GetRequiredIntOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) int64 {
	option := GetRequiredOption(options, key)
	return option.IntValue()
}

func GetBoolOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *bool {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	b := option.BoolValue()
	return &b
}

func GetBoolOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue bool) bool {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.BoolValue()
}

func GetRequiredBoolOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) bool {
	option := GetRequiredOption(options, key)
	return option.BoolValue()
}

func GetUintOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *uint64 {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	u := option.UintValue()
	return &u
}

func GetUintOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue uint64) uint64 {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.UintValue()
}

func GetRequiredUintOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) uint64 {
	option := GetRequiredOption(options, key)
	return option.UintValue()
}

func GetFloatOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) *float64 {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	f := option.FloatValue()
	return &f
}

func GetFloatOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue float64) float64 {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.FloatValue()
}

func GetRequiredFloatOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) float64 {
	option := GetRequiredOption(options, key)
	return option.FloatValue()
}

func GetUserOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) *discordgo.User {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	return option.UserValue(s)
}

func GetUserOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue *discordgo.User, s *discordgo.Session) *discordgo.User {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.UserValue(s)
}

func GetRequiredUserOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) *discordgo.User {
	option := GetRequiredOption(options, key)
	return option.UserValue(s)
}

func GetChannelOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) *discordgo.Channel {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	return option.ChannelValue(s)
}

func GetChannelOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue *discordgo.Channel, s *discordgo.Session) *discordgo.Channel {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.ChannelValue(s)
}

func GetRequiredChannelOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) *discordgo.Channel {
	option := GetRequiredOption(options, key)
	return option.ChannelValue(s)
}

func GetRoleOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session, gid string) *discordgo.Role {
	option := GetOption(options, key)
	if option == nil {
		return nil
	}

	return option.RoleValue(s, gid)
}

func GetRoleOptionOrDefault(options []*discordgo.ApplicationCommandInteractionDataOption, key string, defaultValue *discordgo.Role, s *discordgo.Session, gid string) *discordgo.Role {
	option := GetOption(options, key)
	if option == nil {
		return defaultValue
	}

	return option.RoleValue(s, gid)
}

func GetRequiredRoleOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session, gid string) *discordgo.Role {
	option := GetRequiredOption(options, key)
	return option.RoleValue(s, gid)
}
