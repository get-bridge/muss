package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellVarExpand(t *testing.T) {
	t.Run("unsupported syntax", func(t *testing.T) {
		assert.PanicsWithValue(t, "Invalid interpolation format: '${MUSS_TEST_VAR:+nullorunset}'",
			func() { expand("[${MUSS_TEST_VAR:+nullorunset}]") }, ":+")
	})

	t.Run("var unset", func(t *testing.T) {
		os.Unsetenv("MUSS_TEST_VAR")
		assert.Equal(t, "[]", expand("[$MUSS_TEST_VAR]"), "var")
		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR}]"), "braces")

		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR:-}]"), "default empty")

		assert.Equal(t, "[nullorunset]", expand("[${MUSS_TEST_VAR:-nullorunset}]"), ":-")
		assert.Equal(t, "[unset]", expand("[${MUSS_TEST_VAR-unset}]"), "-")

		assert.PanicsWithValue(t, "Variable 'MUSS_TEST_VAR' is required: nullorunset",
			func() { expand("[${MUSS_TEST_VAR:?nullorunset}]") }, ":?")
		assert.PanicsWithValue(t, "Variable 'MUSS_TEST_VAR' is required: unset",
			func() { expand("[${MUSS_TEST_VAR?unset}]") }, "?")
	})

	t.Run("var blank", func(t *testing.T) {
		os.Setenv("MUSS_TEST_VAR", "")
		assert.Equal(t, "[]", expand("[$MUSS_TEST_VAR]"), "var")
		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR}]"), "braces")

		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR:-}]"), "default empty")

		assert.Equal(t, "[nullorunset]", expand("[${MUSS_TEST_VAR:-nullorunset}]"), ":-")
		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR-unset}]"), "-")

		assert.PanicsWithValue(t, "Variable 'MUSS_TEST_VAR' is required: nullorunset",
			func() { expand("[${MUSS_TEST_VAR:?nullorunset}]") }, ":?")
		assert.Equal(t, "[]", expand("[${MUSS_TEST_VAR?unset}]"), "?")
	})

	t.Run("var nonblank", func(t *testing.T) {
		os.Setenv("MUSS_TEST_VAR", "not blank")
		assert.Equal(t, "[not blank]", expand("[$MUSS_TEST_VAR]"), "var")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR}]"), "braces")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR:-}]"), "default empty")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR:-nullorunset}]"), ":-")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR-unset}]"), "-")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR:?nullorunset}]"), ":?")
		assert.Equal(t, "[not blank]", expand("[${MUSS_TEST_VAR?nullorunset}]"), "?")
	})
}
