package main

import (
	"github.com/galdor/go-program"
)

var buildId string

func main() {
	program := program.NewProgram("emaild", "a self-contained email server")

	program.AddCommand("run", "run the server", cmdRun)
	program.AddCommand("version", "print the version of the server and exit",
		cmdVersion)

	program.ParseCommandLine()
	program.Run()
}
