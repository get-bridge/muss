package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserConfig(t *testing.T) {
	t.Run("user.ToMap", func(t *testing.T) {
		user, err := UserConfigFromMap(map[string]interface{}{
			"override":           map[string]interface{}{},
			"service_preference": []string{"bar"},
			"services": map[string]interface{}{
				"foo": map[string]interface{}{
					"disabled": true,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		m, err := user.ToMap()
		if err != nil {
			t.Fatal(err)
		}

		exp := map[string]interface{}{
			"override":           map[string]interface{}{},
			"service_preference": []string{"bar"},
			"services": map[string]interface{}{
				"foo": map[string]interface{}{
					"config":   "", // zero value
					"disabled": true,
				},
			},
		}

		assert.Equal(t, exp, m)
	})
}
