package smtp

import (
	"net"

	"github.com/galdor/go-log"
)

type ServerConn struct {
	Server *Server
	Log    *log.Logger

	conn net.Conn
}

func (c *ServerConn) Close() {
	c.conn.Close()
}

func (c *ServerConn) main() {
	defer c.Server.wg.Done()

	// TODO
	c.Log.Info("handling connection")

	select {
	case <-c.Server.stopChan:
		return
	}
}
