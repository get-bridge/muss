package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
)

func TestConfigSaveCommand(t *testing.T) {
	t.Run("help description", func(t *testing.T) {
		assert.Equal(t,
			"Generate new docker-compose.yml file.",
			newSaveCommand(nil).Long,
			"default")

		config.SetConfig(map[string]interface{}{"compose_file": "dc.muss.yml"})
		cfg, _ := config.All()

		assert.Equal(t,
			"Generate new dc.muss.yml file.",
			newSaveCommand(cfg).Long,
			"default")
	})
}
