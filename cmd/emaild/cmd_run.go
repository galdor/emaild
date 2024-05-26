package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/galdor/emaild/pkg/server"
	"github.com/galdor/go-log"
	"github.com/galdor/go-program"
)

func cmdRun(p *program.Program) {
	// Command line
	cfgPath := p.OptionValue("cfg")

	// Logger
	logger := log.DefaultLogger("emaild")
	logger.DebugLevel = p.DebugLevel

	// Configuration
	var cfg server.ServerCfg

	if cfgPath != "" {
		logger.Info("loading configuration file %q", cfgPath)

		if err := cfg.Load(cfgPath); err != nil {
			logger.Error("cannot load configuration from %q: %v", cfgPath, err)
			os.Exit(1)
		}
	}

	cfg.Log = logger
	cfg.BuildId = buildId

	// Server
	server := server.NewServer(cfg)

	if err := server.Start(); err != nil {
		logger.Error("cannot start server: %v", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case signo := <-sigChan:
		fmt.Fprintln(os.Stderr)
		logger.Info("received signal %d (%v)", signo, signo)
	}

	server.Stop()
}
