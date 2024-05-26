package smtp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/galdor/go-ejson"
	"github.com/galdor/go-log"
)

type ServerCfg struct {
	Log *log.Logger `json:"-"`

	Host string `json:"host"`
	Port int    `json:"port"`
}

func (cfg *ServerCfg) ValidateJSON(v *ejson.Validator) {
	v.CheckStringNotEmpty("host", cfg.Host)
	v.CheckIntMinMax("port", cfg.Port, 1, 65535)
}

type Server struct {
	Cfg ServerCfg
	Log *log.Logger

	listeners []net.Listener

	conns      []*ServerConn
	connsMutex sync.Mutex

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
	addrs, err := s.resolveHost()
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

func (s *Server) resolveHost() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var resolver net.Resolver

	addrs, err := resolver.LookupHost(ctx, s.Cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve host: %w", err)
	}

	return addrs, nil
}

func (s *Server) Stop() {
	close(s.stopChan)

	s.connsMutex.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.connsMutex.Unlock()

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
	remoteAddr := conn.RemoteAddr().String()

	addr, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return fmt.Errorf("invalid remote address %q: %w", remoteAddr, err)
	}

	s.Log.Debug(1, "accepting connection from %q", addr)

	logData := log.Data{
		"address": addr,
	}

	sconn := ServerConn{
		Server: s,
		Log:    s.Log.Child("conn", logData),

		conn: conn,
	}

	s.connsMutex.Lock()
	s.conns = append(s.conns, &sconn)
	s.connsMutex.Unlock()

	s.wg.Add(1)
	go sconn.main()

	return nil
}
