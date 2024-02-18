package smtp

import (
	"fmt"
	"net"
	"time"

	"github.com/galdor/go-log"
)

const (
	DefaultConnectionTimeout = 10 * time.Second
)

type ClientCfg struct {
	Log               *log.Logger
	ConnectionTimeout time.Duration
}

type Client struct {
	Cfg ClientCfg
	Log *log.Logger

	conn net.Conn
}

func NewClient(address string, cfg ClientCfg) (*Client, error) {
	if cfg.ConnectionTimeout == 0 {
		cfg.ConnectionTimeout = DefaultConnectionTimeout
	}

	conn, err := net.DialTimeout("tcp", address, cfg.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("cannot connect: %w", err)
	}

	c := Client{
		Cfg: cfg,
		Log: cfg.Log,

		conn: conn,
	}

	return &c, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
