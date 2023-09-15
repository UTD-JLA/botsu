package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Database         DatabaseConfig `toml:"database"`
	AoDBPath         string         `toml:"aodb_path"`
	AniDBDumpPath    string         `toml:"anidb_dump_path"`
	VNDBDumpPath     string         `toml:"vndb_dump_path"`
	Token            string         `toml:"token"`
	UseMembersIntent bool           `toml:"use_members_intent"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

func (c *DatabaseConfig) ConnectionURL() url.URL {
	return url.URL{
		Scheme: "postgres",
		Host:   c.Host + fmt.Sprintf(":%d", c.Port),
		User:   url.UserPassword(c.User, c.Password),
		Path:   c.Database,
	}
}

func (c *DatabaseConfig) ConnectionString() string {
	u := c.ConnectionURL()
	return u.String()
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	err = toml.NewDecoder(file).Decode(c)
	if err != nil {
		return err
	}

	return nil
}
