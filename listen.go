package listen

import (
	"net"

	"github.com/inkel/go-proxy-protocol/conn"
)

// WithProxyProtocol returns a net.Listener that will wrap Accept so
// it returns net.Conn that know how to work with Proxy Protocol.
func WithProxyProtocol(network, address string) (net.Listener, error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &listener{ln}, nil
}

type listener struct{ ln net.Listener }

func (l *listener) Close() error   { return l.ln.Close() }
func (l *listener) Addr() net.Addr { return l.ln.Addr() }

func (l *listener) Accept() (net.Conn, error) {
	cn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}

	cn, err = conn.WithProxyProtocol(cn)
	if err != nil {
		return nil, err
	}

	return cn, nil
}
