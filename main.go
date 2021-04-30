package main

import (
	"os"

	"github.com/get-bridge/muss/cmd"
	_ "github.com/get-bridge/muss/cmd/config"
	"github.com/get-bridge/muss/proc"
)

func main() {
	proc.EnableExec()
	os.Exit(cmd.Execute(os.Args[1:]))
}
