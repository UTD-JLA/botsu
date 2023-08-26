package main

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Database DatabaseConfig
	Token    string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c *DatabaseConfig) ConnectionString() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + fmt.Sprintf("%d", c.Port) + "/" + c.Database
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
