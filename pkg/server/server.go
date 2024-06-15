package server

import (
	"fmt"
	"sync"

	"github.com/galdor/emaild/pkg/smtp"
	"github.com/galdor/go-log"
)

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

	smtpServers map[string]*smtp.Server

	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewServer(cfg ServerCfg) (*Server, error) {
	var logger *log.Logger
	if cfg.Logger == nil {
		logger = log.DefaultLogger("emaild")
	} else {
		var err error
		logger, err = log.NewLogger("emaild", *cfg.Logger)
		if err != nil {
			return nil, fmt.Errorf("cannot create logger: %w", err)
		}
	}

	s := Server{
		Cfg: cfg,
		Log: logger,

		smtpServers: make(map[string]*smtp.Server),

		stopChan: make(chan struct{}),
	}

	return &s, nil
}

func (s *Server) Start() error {
	s.Log.Debug(1, "starting")

	if err := s.startSMTPServers(); err != nil {
		return err
	}

	s.Log.Debug(1, "running")
	return nil
}

func (s *Server) startSMTPServers() error {
	for name, pcfg := range s.Cfg.SMTPServers {
		cfg := *pcfg
		cfg.Log = s.Log.Child("smtp_server", log.Data{"server": name})

		server, err := smtp.NewServer(cfg)
		if err != nil {
			return fmt.Errorf("cannot create SMTP server %q: %w", name, err)
		}

		if err := server.Start(); err != nil {
			return fmt.Errorf("cannot start SMTP server %q: %w", err)
		}

		s.smtpServers[name] = server

	}

	return nil
}

func (s *Server) Stop() {
	s.Log.Debug(1, "stopping")

	s.stopSMTPServers()

	close(s.stopChan)
	s.wg.Wait()
}

func (s *Server) stopSMTPServers() {
	for _, server := range s.smtpServers {
		server.Stop()
	}
}
