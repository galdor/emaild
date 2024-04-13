package smtp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/galdor/go-log"
)

type ServerCfg struct {
	Log     *log.Logger
	Address string
}

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

	listener net.Listener

	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewServer(cfg ServerCfg) (*Server, error) {
	c := Server{
		Cfg: cfg,
		Log: cfg.Log,

		stopChan: make(chan struct{}),
	}

	return &c, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Cfg.Address)
	if err != nil {
		return fmt.Errorf("cannot listen on %q: %w", s.Cfg.Address, err)
	}
	s.listener = listener

	s.Log.Info("listening on %q", s.Cfg.Address)

	s.wg.Add(1)
	go s.main()

	return nil
}

func (s *Server) Stop() {
	close(s.stopChan)

	s.listener.Close()

	s.wg.Wait()
}

func (s *Server) main() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			s.Log.Error("cannot accept connection: %v", err)

			select {
			case <-s.stopChan:
				return
			case <-time.After(time.Second):
				continue
			}
		}

		if err := s.handleConnection(conn); err != nil {
			s.Log.Error("%v", err)
			conn.Close()
			continue
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) error {
	remoteAddr := conn.RemoteAddr()
	s.Log.Debug(1, "accepting connection from %q", remoteAddr.String())

	// TODO

	conn.Close()
	return nil
}
