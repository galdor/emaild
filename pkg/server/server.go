package server

import (
	"sync"

	"github.com/galdor/go-log"
)

type ServerCfg struct {
	Log     *log.Logger
	BuildId string
}

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

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

	s.Log.Info("running")
	return nil
}

func (s *Server) Stop() {
	s.Log.Info("stopping")

	close(s.stopChan)
	s.wg.Wait()
}
