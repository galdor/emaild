package main

import (
	"fmt"

	"github.com/galdor/go-program"
)

func cmdVersion(p *program.Program) {
	fmt.Println(buildId)
}
