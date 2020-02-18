package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	assert.Nil(t, QuietErrorOrNil(nil), "or nil")

	err := QuietErrorOrNil(fmt.Errorf("foo"))
	assert.NotNil(t, err)
	assert.Equal(t, "foo", err.Error())
	assert.Equal(t, "foo", errors.Unwrap(err).Error())
}
