package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/galdor/emaild/pkg/imf"
	"github.com/galdor/go-program"
)

func cmdParseMessage(p *program.Program) {
	filePath := p.ArgumentValue("path")

	data, err := readFileOrStdin(filePath)
	if err != nil {
		p.Fatal("%v", err)
	}

	r := imf.NewMessageReader()

	if err := r.Read(data); err != nil {
		p.Fatal("invalid message: %v", err)
	}

	msg, err := r.Close()
	if err != nil {
		p.Fatal("invalid message: %v", err)
	}

	// TODO serialization
	//fmt.Printf("%#v\n", msg)
	for _, field := range msg.Header {
		fmt.Printf("%v\n", field)
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
