package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("no args", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newStartCommand, []string{})

			expOut := `docker-compose
start
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
