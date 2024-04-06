package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/galdor/emaild/pkg/imf"
	"github.com/galdor/go-program"
)

func cmdParseMessage(p *program.Program) {
	outputType := p.OptionValue("output")
	filePath := p.ArgumentValue("path")

	data, err := readFileOrStdin(filePath)
	if err != nil {
		p.Fatal("%v", err)
	}

	d := imf.NewMessageDecoder()
	d.MixedEOL = true

	msg, err := d.DecodeAll(data)
	if err != nil {
		p.Fatal("invalid message: %v", err)
	}

	switch outputType {
	case "encoded":
		e := imf.NewMessageEncoder(msg)

		data, err := e.Encode()
		if err != nil {
			p.Fatal("cannot encode message: %v", err)
		}

		if _, err := os.Stdout.Write(data); err != nil {
			p.Fatal("cannot write stdout: %v", err)
		}

	case "errors":
		var nbErrors int
		for _, field := range msg.Header {
			if field.HasError() {
				nbErrors++
				fmt.Printf("field %q: %s\n", field.Name, field.Error)
			}
		}

		if nbErrors > 0 {
			os.Exit(1)
		}

	case "raw":
		if _, err := os.Stdout.Write(data); err != nil {
			p.Fatal("cannot write stdout: %v", err)
		}

	case "syntax":
		for _, field := range msg.Header {
			fmt.Printf("%v\n", field)
		}

		if len(msg.Body) > 0 {
			fmt.Printf("%v\n", msg.Body)
		}

	default:
		p.Fatal("invalid output type %q", outputType)
	}
}

func readFileOrStdin(filePath string) ([]byte, error) {
	var data []byte
	var err error

	if filePath == "" || filePath == "-" {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("cannot read standard input: %w", err)
		}
	} else {
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("cannot read %q: %w", filePath, err)
		}
	}

	return data, nil
}
