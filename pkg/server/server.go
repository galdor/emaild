package server

import (
	"fmt"
	"sync"

	"github.com/galdor/emaild/pkg/smtp"
	"github.com/galdor/go-log"
)

type ServerCfg struct {
	Log     *log.Logger
	BuildId string
}

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

	smtpServer *smtp.Server

	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewServer(cfg ServerCfg) *Server {

	s := Server{
		Cfg: cfg,
		Log: cfg.Log,

		stopChan: make(chan struct{}),
	}

	return &s
}

func (s *Server) Start() error {
	s.Log.Info("starting")

	smtpServerCfg := smtp.ServerCfg{
		Log:     s.Cfg.Log.Child("smtp_server", nil),
		Address: "127.0.0.1:2525",
	}

	smtpServer, err := smtp.NewServer(smtpServerCfg)
	if err != nil {
		return fmt.Errorf("cannot create smtp server: %w", err)
	}
	s.smtpServer = smtpServer

	if err := s.smtpServer.Start(); err != nil {
		return fmt.Errorf("cannot start smtp server: %w", err)
	}

	s.Log.Info("running")
	return nil
}

func (s *Server) Stop() {
	s.Log.Info("stopping")

	s.smtpServer.Stop()

	close(s.stopChan)
	s.wg.Wait()
}
