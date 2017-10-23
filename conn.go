package viaproxy

import (
	"bufio"
	"bytes"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

// ErrInvalidProxyProtocolHeader is the error returned by Wrap when the proxy
// protocol header is malformed.
var ErrInvalidProxyProtocolHeader = errors.New("invalid proxy protocol header")

// Wrap takes a net.Conn and returns a net.Conn that knows how to
// properly identify the remote address if it comes via a proxy that
// supports the Proxy Protocol.
func Wrap(cn net.Conn) (net.Conn, error) {
	c := &Conn{Conn: cn, r: bufio.NewReader(cn)}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

// Conn is an implementation of net.Conn interface for TCP connections that come
// from a proxy that users the Proxy Protocol to communicate with the upstream
// servers.
type Conn struct {
	net.Conn
	r       *bufio.Reader
	proxy   net.Addr
	remote  net.Addr
	proxied bool
}

// ProxyAddr returns the proxy remote network address.
func (c *Conn) ProxyAddr() net.Addr { return c.proxy }

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.remote != nil {
		return c.remote
	}
	return c.Conn.RemoteAddr()
}

// Read reads data from the connection.
func (c *Conn) Read(b []byte) (int, error) { return c.r.Read(b) }

func (c *Conn) init() error {
	unknown := []byte("PROXY UNKNOWN\r\n")
	buf, err := c.r.Peek(len(unknown))
	if err != nil {
		return errors.Wrap(err, "parsing proxy protocol header")
	}
	if bytes.Equal(buf, unknown) {
		_, err = c.r.Discard(len(unknown))
		return err
	}

	// PROXY
	buf = make([]byte, 6)
	_, err = c.r.Read(buf)
	if err != nil {
		return err
	}
	if !bytes.Equal(buf, []byte("PROXY ")) {
		return errors.Errorf("invalid proxy protocol header prefix: %v", buf)
	}

	// TCP4 || TCP6
	buf = make([]byte, 5)
	_, err = c.r.Read(buf)
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header")
	}
	if !bytes.Equal([]byte("TCP4 "), buf) && !bytes.Equal([]byte("TCP6 "), buf) {
		return errors.Errorf("unrecognized protocol: %q", buf)
	}

	// CLIENT IP
	p, err := c.r.ReadString(' ')
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header while reading client ip")
	}
	clientIP := net.ParseIP(p[:len(p)-1])
	if clientIP == nil {
		return errors.Errorf("cannot parse client ip %q", p)
	}

	// PROXY IP
	p, err = c.r.ReadString(' ')
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header while reading proxyt ip")
	}
	proxyIP := net.ParseIP(p[:len(p)-1])
	if proxyIP == nil {
		return errors.Errorf("cannot parse proxy ip %q", p)
	}

	// CLIENT PORT
	p, err = c.r.ReadString(' ')
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header while reading client port")
	}
	clientPort, err := strconv.Atoi(p[:len(p)-1])
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header parsing client port")
	}

	// PROXY PORT
	p, err = c.r.ReadString('\r')
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header while reading proxy port")
	}
	proxyPort, err := strconv.Atoi(p[:len(p)-1])
	if err != nil {
		return errors.Wrap(err, "invalid proxy protocol header parsing proxy port")
	}

	// Trailing
	b, err := c.r.ReadByte()
	if err != nil || b != '\n' {
		return errors.Wrap(err, "invalid trailing")
	}

	c.remote = &net.TCPAddr{IP: clientIP, Port: clientPort}
	c.proxy = &net.TCPAddr{IP: proxyIP, Port: proxyPort}

	return nil
}
