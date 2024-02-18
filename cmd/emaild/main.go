package main

import (
	"github.com/galdor/go-program"
)

var buildId string

func main() {
	program := program.NewProgram("emaild", "a self-contained email server")

	program.AddCommand("run", "run the server", cmdRun)

	program.ParseCommandLine()
	program.Run()
}
