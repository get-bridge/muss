package main

import (
	"os"

	"github.com/instructure-bridge/muss/cmd"
	_ "github.com/instructure-bridge/muss/cmd/config"
	"github.com/instructure-bridge/muss/proc"
)

func main() {
	proc.EnableExec()
	os.Exit(cmd.Execute(os.Args[1:]))
}
