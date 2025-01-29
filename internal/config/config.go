package config

import (
	"fmt"
	"runtime"
)

// Config holds the generator configuration
type Config struct {
	Pattern       string
	Position      string
	Threads       int
	CaseSensitive bool
	OutputFile    string
	Format        string
	Quiet         bool
	Count         int
	Stats         bool
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate position
	validPositions := map[string]bool{
		"start": true,
		"end":   true,
		"any":   true,
	}
	if !validPositions[c.Position] {
		return fmt.Errorf("invalid position '%s': must be one of: start, end, any", c.Position)
	}

	// Validate pattern length
	if len(c.Pattern) == 0 {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Validate threads
	if c.Threads < 1 {
		return fmt.Errorf("number of threads must be at least 1")
	}

	// Validate count
	if c.Count < 1 {
		return fmt.Errorf("count must be at least 1")
	}

	// Validate format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validFormats[c.Format] {
		return fmt.Errorf("invalid format '%s': must be one of: text, json", c.Format)
	}

	return nil
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Position:      "end",
		Threads:       runtime.NumCPU(),
		CaseSensitive: false,
		Format:        "text",
		Count:         1,
	}
}
