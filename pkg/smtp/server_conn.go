package smtp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/galdor/emaild/pkg/utils"
	"github.com/galdor/go-log"
)

type ExpectedError struct {
	Err error
}

func NewExpectedError(err error) *ExpectedError {
	return &ExpectedError{Err: err}
}

func (err *ExpectedError) Error() string {
	return err.Err.Error()
}

func (err *ExpectedError) Unwrap() error {
	return err.Err
}

type ServerConn struct {
	Server *Server
	Log    *log.Logger

	domain string // value sent by EHLO or HELO

	conn net.Conn
	rbuf *bufio.Reader
	wbuf bytes.Buffer
}

func (c *ServerConn) Start() {
	c.rbuf = bufio.NewReader(c.conn)

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

	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				var expectedError *ExpectedError

				if errors.As(err, &expectedError) {
					return
				}
			}

			msg := utils.RecoverValueString(v)
			trace := utils.StackTrace(0, 20, true)

			c.Log.Error("panic: %s\n%s", msg, trace)
		}
	}()

	c.writeGreeting()

	for {
		r, err := c.readRequest()
		if err != nil {
			c.Log.Error("cannot read request: %v", err)
			return
		}

		if err := c.processRequest(r); err != nil {
			c.Log.Error("%s: %v", r.Keyword, err)
			return
		}
	}

	select {
	case <-c.Server.stopChan:
		return
	}
}

func (c *ServerConn) readLine() ([]byte, error) {
	s, err := c.rbuf.ReadBytes('\n')
	if err != nil {
		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			panic(NewExpectedError(err))
		}

		return nil, fmt.Errorf("cannot read connection: %w", err)
	}

	if len(s) < 2 || s[len(s)-2] != '\r' {
		return nil, fmt.Errorf("missing carriage return before newline")
	}

	return s[:len(s)-2], nil
}

func (c *ServerConn) readRequest() (*LineReader, error) {
	line, err := c.readLine()
	if err != nil {
		return nil, fmt.Errorf("cannot read line: %w", err)
	}

	return NewLineReader(line)
}

func (c *ServerConn) writeLine(code int, more bool, line string) {
	c.wbuf.Reset()

	fmt.Fprintf(&c.wbuf, "%d", code)

	if more {
		c.wbuf.WriteByte('-')
	} else if len(line) > 0 {
		c.wbuf.WriteByte(' ')
	}

	c.wbuf.WriteString(line)

	c.wbuf.WriteString("\r\n")

	if _, err := io.Copy(c.conn, &c.wbuf); err != nil {
		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			panic(NewExpectedError(err))
		}

		panic(err)
	}
}

func (c *ServerConn) writeError(code int, format string, args ...any) {
	c.writeLine(code, false, fmt.Sprintf(format, args...))
}

func (c *ServerConn) writeGreeting() {
	c.writeLine(220, false, c.Server.Cfg.PublicHost)
}

func (c *ServerConn) processRequest(r *LineReader) error {
	var fn func(*LineReader) error

	switch r.Keyword {
	case "EHLO":
		fn = c.processEHLO
	case "HELO":
		fn = c.processHELO
	case "RSET":
		fn = c.processRSET
	default:
		return fmt.Errorf("unknown keyword %q", r.Keyword)
	}

	return fn(r)
}

func (c *ServerConn) processEHLO(r *LineReader) error {
	domainData := r.ReadAll()

	domain, err := ValidateDomain(domainData)
	if err != nil {
		c.writeError(501, "invalid domain: %v", err)
		return fmt.Errorf("invalid domain %q: %w", domainData, err)
	}

	c.domain = domain

	extensions := c.Server.extensions

	c.writeLine(250, len(extensions) > 0, c.Server.Cfg.PublicHost)

	i := 0
	for name, value := range extensions {
		line := name
		if value != "" {
			line += " " + value
		}

		c.writeLine(250, i < len(extensions)-1, line)
		i++
	}

	// RFC 5321 4.1.4.
	c.reset()

	return nil
}

func (c *ServerConn) processHELO(r *LineReader) error {
	domainData := r.ReadUntilWhitespace()

	domain, err := ValidateDomain(domainData)
	if err != nil {
		c.writeError(501, "invalid domain: %v", err)
		return fmt.Errorf("invalid domain %q: %w", domainData, err)
	}

	c.domain = domain

	c.writeLine(250, false, c.Server.Cfg.PublicHost)

	// RFC 5321 4.1.4.
	c.reset()

	return nil
}

func (c *ServerConn) processRSET(r *LineReader) error {
	c.reset()
	return nil
}

func (c *ServerConn) reset() {
	// TODO
}
