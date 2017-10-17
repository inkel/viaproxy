package viaproxy

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
)

// ErrInvalidProxyProtocolHeader is the error returned by Wrap when the proxy
// protocol header is malformed.
var ErrInvalidProxyProtocolHeader = errors.New("invalid proxy protocol header")

// Wrap takes a net.Conn and returns a net.Conn that knows how to
// properly identify the remote address if it comes via a proxy that
// supports the Proxy Protocol.
func Wrap(cn net.Conn) (net.Conn, error) {
	c := &conn{Conn: cn, r: bufio.NewReader(cn)}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

type conn struct {
	net.Conn
	r       *bufio.Reader
	local   net.Addr
	remote  net.Addr
	proxied bool
}

func (c *conn) LocalAddr() net.Addr  { return c.local }
func (c *conn) RemoteAddr() net.Addr { return c.remote }

func (c *conn) Read(b []byte) (int, error) { return c.r.Read(b) }

func (c *conn) init() error {
	c.local = c.Conn.LocalAddr()
	c.remote = c.Conn.RemoteAddr()

	buf, err := c.r.Peek(6)
	if err != io.EOF && err != nil {
		return err
	}

	if err == io.EOF {
		return nil
	}

	if !bytes.Equal([]byte("PROXY "), buf) {
		return nil
	}

	c.proxied = true
	proxyLine, err := c.r.ReadString('\n')
	if err != nil {
		return err
	}
	fields := strings.Fields(proxyLine)

	if len(fields) == 2 && fields[1] == "UNKNOWN" {
		return nil
	}

	if len(fields) != 6 {
		return ErrInvalidProxyProtocolHeader
	}

	clientIP := net.ParseIP(fields[2])
	clientPort, err := strconv.Atoi(fields[4])
	if clientIP == nil || err != nil {
		return ErrInvalidProxyProtocolHeader
	}

	proxyIP := net.ParseIP(fields[3])
	proxyPort, err := strconv.Atoi(fields[5])
	if proxyIP == nil || err != nil {
		return ErrInvalidProxyProtocolHeader
	}

	c.remote = &net.TCPAddr{IP: clientIP, Port: clientPort}
	c.local = &net.TCPAddr{IP: proxyIP, Port: proxyPort}

	return nil
}
