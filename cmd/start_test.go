package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("no args", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"start"})

			expOut := `docker-compose
start
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
