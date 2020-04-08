package config

// Warn adds a warning to the config (which the commands will print)
// (as long as the message hasn't already been Warn()ed).
func (cfg *ProjectConfig) Warn(msg string) {
	for _, w := range cfg.Warnings {
		if msg == w {
			return
		}
	}
	cfg.Warnings = append(cfg.Warnings, msg)
}
