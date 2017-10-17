package viaproxy_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"testing"

	"github.com/inkel/viaproxy"
)

type conn struct {
	net.Conn
	data io.Reader
}

func (c *conn) Read(b []byte) (n int, err error) { return c.data.Read(b) }

func (c *conn) LocalAddr() net.Addr  { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9876} }
func (c *conn) RemoteAddr() net.Addr { return &net.TCPAddr{IP: net.ParseIP("10.0.1.2"), Port: 1234} }

func testConn(data []byte) net.Conn { return &conn{data: bytes.NewReader(data)} }

type testAddr string

func (a testAddr) Network() string { return "tcp" }
func (a testAddr) String() string  { return string(a) }

func equalAddr(a, b net.Addr) bool {
	return a.Network() == b.Network() && a.String() == b.String()
}
func TestWrap(t *testing.T) {
	cases := []struct {
		line, data []byte
		remoteAddr testAddr
		err        error
	}{
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 5678 1234\r\nfoo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "192.168.1.20:5678", nil},
		{[]byte("PROXY UNKNOWN\r\nfoo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "10.0.1.2:1234", nil},

		{[]byte("foo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "10.0.1.2:1234", nil},
		{[]byte("\x00\r\n"), []byte("\x00\r\n"), "10.0.1.2:1234", nil},

		// Invalid proxy protocol lines
		{[]byte("PROXY TCP5\r\n"), nil, "", viaproxy.ErrInvalidProxyProtocolHeader},
		{[]byte("PROXY TCP4 192.168.X.20 10.0.0.1 5678 1234\r\n"), nil, "", viaproxy.ErrInvalidProxyProtocolHeader},
		{[]byte("PROXY TCP4 192.168.1.20 10.X.0.1 5678 1234\r\n"), nil, "", viaproxy.ErrInvalidProxyProtocolHeader},
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 567X 1234\r\n"), nil, "", viaproxy.ErrInvalidProxyProtocolHeader},
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 5678 123X\r\n"), nil, "", viaproxy.ErrInvalidProxyProtocolHeader},
	}

	for _, c := range cases {
		t.Run(string(c.line), func(t *testing.T) {
			cn := testConn(c.line)

			cn, err := viaproxy.Wrap(cn)
			if err != c.err {
				t.Fatalf("expecting error %v, got %v", c.err, err)
			}
			if cn == nil && c.err != nil {
				// no need to continue processing
				return
			}

			if !equalAddr(c.remoteAddr, cn.RemoteAddr()) {
				t.Errorf("expecting RemoteAddr() %v, got %v", c.remoteAddr, cn.RemoteAddr())
			}

			data, err := ioutil.ReadAll(cn)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(c.data, data) {
				t.Errorf("expecting data %q, got %q", c.data, data)
			}
		})
	}
}
