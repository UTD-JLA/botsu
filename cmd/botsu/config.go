package main

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"net/url"
	"os"
	"slices"
	"strings"
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

	// used to set connection string, ignoring
	// the other properties
	urlOverride *url.URL `toml:"-"`
}

func (c *DatabaseConfig) ConnectionURL() url.URL {
	if c.urlOverride != nil {
		return *c.urlOverride
	}

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

func (c *Config) LoadDefaults() {
	if c.VNDBDumpPath == "" {
		c.VNDBDumpPath = "data/vndb-db"
	}

	if c.AniDBDumpPath == "" {
		c.AniDBDumpPath = "data/anime-titles.dat"
	}

	if c.AoDBPath == "" {
		c.AoDBPath = "data/anime-offline-database-minified.json"
	}
}

func (c *Config) LoadEnv() error {
	token, ok := os.LookupEnv("BOTSU_TOKEN")

	if ok {
		c.Token = token
	}

	aodbPath, ok := os.LookupEnv("BOTSU_AODB_PATH")

	if ok {
		c.AoDBPath = aodbPath
	}

	anidbPath, ok := os.LookupEnv("BOTSU_ANIDB_PATH")

	if ok {
		c.AniDBDumpPath = anidbPath
	}

	vndbPath, ok := os.LookupEnv("BOTSU_VNDB_PATH")

	if ok {
		c.VNDBDumpPath = vndbPath
	}

	connectionString, ok := os.LookupEnv("BOTSU_CONNECTION_STRING")

	if ok {
		connectionURL, err := url.Parse(connectionString)

		if err != nil {
			return err
		}

		c.Database.urlOverride = connectionURL
	}

	useMembersIntent, ok := os.LookupEnv("BOTSU_USE_MEMBERS_INTENT")

	if ok {
		c.UseMembersIntent = slices.Contains([]string{
			"t",
			"true",
			"1",
		}, strings.ToLower(useMembersIntent))
	}

	return nil
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
