package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/get-bridge/muss/cmd"
	"github.com/get-bridge/muss/config"
)

func assertHasSubCommand(t *testing.T, args ...string) {
	cfg, _ := config.NewConfigFromMap(nil)
	rootCmd := cmd.NewRootCommand(cfg)
	found, _, err := rootCmd.Find(args)
	assert.Nil(t, err)
	assert.NotNil(t, found)
}

func TestMain(t *testing.T) {
	assertHasSubCommand(t, "attach")

	// Prove that "main" loads all the subcommand packages.
	assertHasSubCommand(t, "config")
	assertHasSubCommand(t, "config", "show")
}
