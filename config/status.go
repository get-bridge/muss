package config

import (
	"time"
)

// StatusConfig defines how to determine project status.
type StatusConfig struct {
	Exec       []string      `yaml:"exec"`
	LineFormat string        `yaml:"line_format"`
	Interval   time.Duration `yaml:"interval"`
}
