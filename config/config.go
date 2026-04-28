// Package config provides application configuration management.
package config

import (
	"fmt"
	"strconv"
	"sync"
)

const (
	defaultRetries = 3
	defaultTimeout = "5s"
	defaultLogLevel = "info"
)

// Option keys.
const (
	KeyRetries  = "retries"
	KeyTimeout  = "timeout"
	KeyLogLevel = "log_level"
	KeyDebug    = "debug"
)

// Config holds arbitrary key-value configuration options.
type Config struct {
	mu      sync.RWMutex
	Options map[string]interface{}
}

// New creates a Config pre-populated with defaults.
func New() *Config {
	return &Config{
		Options: map[string]interface{}{
			KeyRetries:  defaultRetries,
			KeyTimeout:  defaultTimeout,
			KeyLogLevel: defaultLogLevel,
			KeyDebug:    false,
		},
	}
}

// Set stores a value for the given key (thread-safe).
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Options[key] = value
}

// Get retrieves the value for the given key (thread-safe).
func (c *Config) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Options[key]
	return v, ok
}

// GetString returns the value for key as a string, or the fallback.
func (c *Config) GetString(key, fallback string) string {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt returns the value for key as an int, or the fallback.
func (c *Config) GetInt(key string, fallback int) int {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	switch t := v.(type) {
	case int:
		return t
	case string:
		n, err := strconv.Atoi(t)
		if err != nil {
			return fallback
		}
		return n
	default:
		return fallback
	}
}

// GetBool returns the value for key as a bool, or the fallback.
func (c *Config) GetBool(key string, fallback bool) bool {
	v, ok := c.Get(key)
	if !ok {
		return fallback
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}

// Merge copies all entries from other into c, overwriting duplicates.
func (c *Config) Merge(other *Config) {
	other.mu.RLock()
	defer other.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range other.Options {
		c.Options[k] = v
	}
}

// Summary returns a human-readable dump of the configuration.
func (c *Config) Summary() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("Config{retries:%v, timeout:%v, log_level:%v, debug:%v}",
		c.Options[KeyRetries],
		c.Options[KeyTimeout],
		c.Options[KeyLogLevel],
		c.Options[KeyDebug],
	)
}
