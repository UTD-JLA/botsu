package data

import "time"

const (
	defaultMaxBuffLen = 1000
	defaultMaxBuffAge = time.Second
)

type storeConfig struct {
	Path         string
	SearchFields []string
	MaxBuffLen   int
	MaxBuffAge   time.Duration
}

func (c *storeConfig) applyDefaults() {
	if c.MaxBuffLen == 0 {
		c.MaxBuffLen = defaultMaxBuffLen
	}

	if c.MaxBuffAge == 0 {
		c.MaxBuffAge = defaultMaxBuffAge
	}
}

func (c *storeConfig) applyOptions(opts []Option) {
	for _, opt := range opts {
		opt(c)
	}
}

type Option func(*storeConfig)

func WithPath(path string) Option {
	return func(c *storeConfig) {
		c.Path = path
	}
}

func WithSearchFields(fields ...string) Option {
	return func(c *storeConfig) {
		c.SearchFields = fields
	}
}

func WithMaxBuffLen(len int) Option {
	return func(c *storeConfig) {
		c.MaxBuffLen = len
	}
}

func WithMaxBuffAge(age time.Duration) Option {
	return func(c *storeConfig) {
		c.MaxBuffAge = age
	}
}
