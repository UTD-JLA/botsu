package data

import "time"

const (
	defaultMaxBuffLen = 1000
	defaultMaxBuffAge = time.Second
)

type StoreConfig struct {
	Path         string
	SearchFields []string
	MaxBuffLen   int
	MaxBuffAge   time.Duration
}

func (c StoreConfig) WithPath(path string) StoreConfig {
	c.Path = path
	return c
}

func (c StoreConfig) WithSearchFields(fields ...string) StoreConfig {
	c.SearchFields = fields
	return c
}

func (c StoreConfig) WithMaxBuffLen(len int) StoreConfig {
	c.MaxBuffLen = len
	return c
}

func (c StoreConfig) WithMaxBuffAge(age time.Duration) StoreConfig {
	c.MaxBuffAge = age
	return c
}

func NewDefaultConfig(path string) StoreConfig {
	return StoreConfig{
		Path:         path,
		SearchFields: []string{},
		MaxBuffLen:   defaultMaxBuffLen,
		MaxBuffAge:   defaultMaxBuffAge,
	}
}
