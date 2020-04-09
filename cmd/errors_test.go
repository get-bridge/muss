package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.Nil(t, QuietErrorOrNil(nil), "or nil")

	err := QuietErrorOrNil(errors.New("foo"))
	assert.NotNil(t, err)
	assert.Equal(t, "foo", err.Error())
	assert.Equal(t, "foo", errors.Unwrap(err).Error())
}
