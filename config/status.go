package config

import (
	"time"
)

// StatusConfig defines how to determine project status.
type StatusConfig struct {
	Exec       []string      `mapstructure:"exec"`
	LineFormat string        `mapstructure:"line_format"`
	Interval   time.Duration `mapstructure:"interval"`
}
