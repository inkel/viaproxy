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
	return (a == nil && b == nil) || a.Network() == b.Network() && a.String() == b.String()
}

func TestWrap(t *testing.T) {
	cases := []struct {
		line, data []byte
		remoteIP   string
		remotePort int
		proxy      net.Addr
		err        bool
	}{
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 5678 1234\r\nfoo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "192.168.1.20", 5678, &net.TCPAddr{IP: net.IPv4(10, 0, 0, 1), Port: 1234}, false},
		{[]byte("PROXY TCP6 fe80::aede:48ff:fe00:1122 ::1 5678 1234\r\nfoo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "fe80::aede:48ff:fe00:1122", 5678, &net.TCPAddr{IP: net.ParseIP("::1"), Port: 1234}, false},
		{[]byte("PROXY UNKNOWN\r\nfoo\r\nbar\r\n"), []byte("foo\r\nbar\r\n"), "10.0.1.2", 1234, nil, false},

		// Invalid proxy protocol lines
		{[]byte("GET / HTTP/1.0\r\n"), nil, "", -1, nil, true},
		{[]byte("PROXY TCP5\r\n"), nil, "", -1, nil, true},
		{[]byte("PROXY TCP4 192.168.X.20 10.0.0.1 5678 1234\r\n"), nil, "", -1, nil, true},
		{[]byte("PROXY TCP4 192.168.1.20 10.X.0.1 5678 1234\r\n"), nil, "", -1, nil, true},
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 567X 1234\r\n"), nil, "", -1, nil, true},
		{[]byte("PROXY TCP4 192.168.1.20 10.0.0.1 5678 123X\r\n"), nil, "", -1, nil, true},
	}

	for _, c := range cases {
		t.Run(string(c.line), func(t *testing.T) {
			cn := testConn(c.line)

			cn, err := viaproxy.Wrap(cn)
			if c.err && err == nil {
				t.Fatal("expecting error, got nil")
			}
			if !c.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cn == nil && c.err && err != nil {
				// no need to continue processing
				return
			}

			var remote net.Addr = &net.TCPAddr{IP: net.ParseIP(c.remoteIP), Port: c.remotePort}
			if !equalAddr(remote, cn.RemoteAddr()) {
				t.Errorf("expecting RemoteAddr() %v, got %v", remote, cn.RemoteAddr())
			}

			data, err := ioutil.ReadAll(cn)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(c.data, data) {
				t.Errorf("expecting data %q, got %q", c.data, data)
			}

			pcn, ok := cn.(*viaproxy.Conn)
			if !ok {
				t.Fatalf("cannot cast connection to *viaproxy.Conn")
			}
			if !equalAddr(pcn.ProxyAddr(), c.proxy) {
				t.Errorf("expecting ProxyAddr() %v, got %v", c.proxy, pcn.ProxyAddr())
			}
		})
	}
}
