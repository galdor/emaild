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
	logger := log.DefaultLogger("emaild")

	cfg := server.ServerCfg{
		Log:     logger,
		BuildId: buildId,
	}

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
