package discordutil

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var OptionNotFound = errors.New("option not found")

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

func GetRequiredOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (*discordgo.ApplicationCommandInteractionDataOption, error) {
	if option := GetOption(options, key); option != nil {
		return option, nil
	}

	return nil, fmt.Errorf("%w: %s", OptionNotFound, key)
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

func GetRequiredStringOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (string, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return "", err
	}

	return option.StringValue(), nil
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

func GetRequiredIntOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (int64, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return 0, err
	}

	return option.IntValue(), nil
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

func GetRequiredBoolOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (bool, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return false, err
	}

	return option.BoolValue(), nil
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

func GetRequiredUintOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (uint64, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return 0, err
	}

	return option.UintValue(), nil
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

func GetRequiredFloatOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string) (float64, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return 0, err
	}

	return option.FloatValue(), nil
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

func GetRequiredUserOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) (*discordgo.User, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return nil, err
	}

	return option.UserValue(s), nil
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

func GetRequiredChannelOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session) (*discordgo.Channel, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return nil, err
	}

	return option.ChannelValue(s), nil
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

func GetRequiredRoleOption(options []*discordgo.ApplicationCommandInteractionDataOption, key string, s *discordgo.Session, gid string) (*discordgo.Role, error) {
	option, err := GetRequiredOption(options, key)
	if err != nil {
		return nil, err
	}

	return option.RoleValue(s, gid), nil
}
