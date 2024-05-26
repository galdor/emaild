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

func (c *ServerConn) Start() {
	c.Server.wg.Add(1)
	go c.main()
}

func (c *ServerConn) Close() {
	c.conn.Close()
}

func (c *ServerConn) main() {
	defer func() {
		c.conn.Close()

		c.Server.connsMutex.Lock()
		delete(c.Server.conns, c)
		c.Server.connsMutex.Unlock()

		c.Server.wg.Done()
	}()

	// TODO
	c.Log.Debug(1, "handling connection")

	select {
	case <-c.Server.stopChan:
		return
	}
}

func (c *ServerConn) readLine() (string, error) {
	// TODO
	return "", nil
}
