package proc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExec(t *testing.T) {
	t.Run("stub", func(t *testing.T) {
		LastExecArgv = nil
		err := Exec([]string{"echo", "hoge", "piyo"})
		assert.Nil(t, err)
		assert.Equal(t, LastExecArgv, []string{"echo", "hoge", "piyo"})
	})
}
