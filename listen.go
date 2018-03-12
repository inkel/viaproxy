package viaproxy

import (
	"net"
)

// Listen returns a net.Listener that will wrap Accept so it returns
// net.Conn that know how to work with Proxy Protocol.
func Listen(network, address string) (*Listener, error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &Listener{ln}, nil
}

// Listener is a wrap on net.Listener that returns wrapped Conn
// objects.
type Listener struct{ ln net.Listener }

// Close stops listening on the TCP address. Already Accepted
// connections are not closed.
func (l *Listener) Close() error { return l.ln.Close() }

// Addr returns the listener's network address, a *TCPAddr. The Addr
// returned is shared by all invocations of Addr, so do not modify it.
func (l *Listener) Addr() net.Addr { return l.ln.Addr() }

// Accept implements the Accept method in the Listener interface; it
// waits for the next call and returns a generic Conn.
func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptFromProxy()
}

// AcceptFromProxy accepts the next incoming call and returns the new
// connection.
func (l *Listener) AcceptFromProxy() (*Conn, error) {
	cn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	return Wrap(cn)
}
