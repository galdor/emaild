package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/galdor/emaild/pkg/server"
	"github.com/galdor/go-program"
)

func cmdRun(p *program.Program) {
	// Command line
	cfgPath := p.OptionValue("cfg")

	// Configuration
	var cfg server.ServerCfg

	if cfgPath != "" {
		p.Info("loading configuration file %q", cfgPath)

		if err := cfg.Load(cfgPath); err != nil {
			p.Fatal("cannot load configuration from %q: %v", cfgPath, err)
		}
	}

	cfg.BuildId = buildId

	// Server
	server, err := server.NewServer(cfg)
	if err != nil {
		p.Fatal("cannot create server: %v", err)
	}

	if err := server.Start(); err != nil {
		p.Fatal("cannot start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case signo := <-sigChan:
		fmt.Fprintln(os.Stderr)
		p.Info("received signal %d (%v)", signo, signo)
	}

	server.Stop()
}
