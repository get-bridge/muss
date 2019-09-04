package main

import (
	"os"

	"gerrit.instructure.com/muss/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
