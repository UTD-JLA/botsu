package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Database DatabaseConfig `toml:"database"`
	AodbPath string         `toml:"aodb_path"`
	Token    string         `toml:"token"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

func (c *DatabaseConfig) ConnectionString() string {
	connectionUrl := url.URL{
		Scheme: "postgres",
		Host:   c.Host + fmt.Sprintf(":%d", c.Port),
		User:   url.UserPassword(c.User, c.Password),
		Path:   c.Database,
	}

	return connectionUrl.String()
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
