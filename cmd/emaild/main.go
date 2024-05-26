package main

import (
	"github.com/galdor/go-program"
)

var buildId string

func main() {
	var c *program.Command

	program := program.NewProgram("emaild", "a self-contained email server")

	c = program.AddCommand("parse-message", "parse a message", cmdParseMessage)
	c.AddOption("o", "output", "type", "raw", "the type of output: "+
		"encoded, errors, raw, syntax")
	c.AddOptionalArgument("path", "the path of the message file")

	c = program.AddCommand("run", "run the server", cmdRun)
	c.AddOption("c", "cfg", "path", "", "the path of the configuration file")

	program.AddCommand("version", "print the version of the server and exit",
		cmdVersion)

	program.ParseCommandLine()
	program.Run()
}
