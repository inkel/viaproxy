package viaproxy

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// ErrInvalidProxyProtocolHeader is the error returned by Wrap when the proxy
// protocol header is malformed.
var ErrInvalidProxyProtocolHeader = errors.New("invalid proxy protocol header")

// Wrap takes a net.Conn and returns a pointer to Conn that knows how to
// properly identify the remote address if it comes via a proxy that
// supports the Proxy Protocol.
func Wrap(cn net.Conn) (*Conn, error) {
	c := &Conn{cn: cn, r: bufio.NewReader(cn)}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

// Conn is an implementation of net.Conn interface for TCP connections that come
// from a proxy that users the Proxy Protocol to communicate with the upstream
// servers.
type Conn struct {
	cn     net.Conn
	r      *bufio.Reader
	proxy  net.Addr
	remote net.Addr
}

// ProxyAddr returns the proxy remote network address.
func (c *Conn) ProxyAddr() net.Addr { return c.proxy }

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.remote != nil {
		return c.remote
	}
	return c.cn.RemoteAddr()
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr { return c.cn.LocalAddr() }

// Read reads data from the connection.
func (c *Conn) Read(b []byte) (int, error) { return c.r.Read(b) }

// Close closes the connection.
func (c *Conn) Close() error { return c.cn.Close() }

// SetDeadline implements the Conn SetDeadline method.
func (c *Conn) SetDeadline(t time.Time) error { return c.cn.SetDeadline(t) }

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *Conn) SetReadDeadline(t time.Time) error { return c.cn.SetReadDeadline(t) }

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *Conn) SetWriteDeadline(t time.Time) error { return c.cn.SetWriteDeadline(t) }

// Write implements the Conn Write method.
func (c *Conn) Write(b []byte) (int, error) { return c.cn.Write(b) }

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
