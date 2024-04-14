package smtp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/galdor/go-log"
)

type ServerCfg struct {
	Log *log.Logger

	Hosts []string
	Port  int
}

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

	listeners []net.Listener

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
	addrs, err := s.resolveHosts(s.Cfg.Hosts)
	if err != nil {
		return err
	}

	addrTable := make(map[string]struct{})
	for _, addr := range addrs {
		addrTable[addr] = struct{}{}
	}

	port := strconv.Itoa(s.Cfg.Port)

	for addr := range addrTable {
		listener, err := net.Listen("tcp", net.JoinHostPort(addr, port))
		if err != nil {
			return fmt.Errorf("cannot listen on %q: %w", addr, err)
		}

		s.Log.Info("listening on %q", addr)

		s.listeners = append(s.listeners, listener)

		s.wg.Add(1)
		go s.listen(listener)
	}

	return nil
}

func (s *Server) resolveHosts(hosts []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var resolver net.Resolver
	var addrs []string

	for _, host := range hosts {
		hAddrs, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve host %q: %w", host, err)
		}

		addrs = append(addrs, hAddrs...)
	}

	return addrs, nil
}

func (s *Server) Stop() {
	close(s.stopChan)

	for _, listener := range s.listeners {
		listener.Close()
	}

	s.wg.Wait()
}

func (s *Server) listen(listener net.Listener) {
	defer s.wg.Done()

	for {
		conn, err := listener.Accept()
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
