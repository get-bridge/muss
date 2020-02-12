package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogsCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"logs",
				"--no-color",
				"-f",
				"-t",
				"--tail=5",
				"app",
				"web",
			})

			// sorted and normalized
			expOut := `docker-compose
logs
--follow
--no-color
--tail=5
--timestamps
app
web
`
			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("some args", func(*testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"logs",
				"-f",
				"--tail", "all",
				"app",
				"web",
			})

			expOut := `docker-compose
logs
--follow
--tail=all
app
web
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("no args", func(*testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"logs"})

			expOut := `docker-compose
logs
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
