# Proxy Protocol support for Go net.Conn

[![GoDoc](https://godoc.org/github.com/inkel/viaproxy?status.svg)](https://godoc.org/github.com/inkel/viaproxy) [![Go Report Card](https://goreportcard.com/badge/github.com/inkel/viaproxy)](https://goreportcard.com/report/github.com/inkel/viaproxy)

Regular Go `net` doesn't support [Proxy Protocol](http://www.haproxy.com/blog/haproxy/proxy-protocol/) when being load balanced with this option enabled. This makes you loose the original remote address and will report the load balancer's address instead on `net.Conn.RemoteAddr()`.  This package adds allows you to create `net.Conn` objects that know how to understand Proxy Protocol.

You can read more about this in my [Proxy Protocol: what is it and how to use it with Go](https://inkel.github.io/posts/proxy-protocol/) article.

## Usage
In your server, you can do the following:

```go
ln, err := net.Listen("tcp", *addr)
if err != nil {
	log.Fatal(err)
}

for {
	cn, err := ln.Accept()
	if err != nil {
		log.Println("ln.Accept():", err)
		continue
	}

	pcn, err := viaproxy.Wrap(cn)
	if err != nil {
		log.Println("Wrap():", err)
		continue
	}

	log.Printf("remote address is: %v", pcn.RemoteAddr())
	log.Printf("local address is: %v", pcn.LocalAddr())
	log.Printf("proxy address is: %v", pcn.ProxyAddr())
	pcn.Close()
}
```

Given that one can forget about this, you can also do the following:

```go
ln, err := viaproxy.Listen("tcp", *addr)
if err != nil {
	log.Fatal(err)
}

for {
	cn, err := ln.Accept()
	if err != nil {
		log.Println("ln.Accept():", err)
		continue
	}

	// The connection should be safe to be converted to a *viaproxy.Conn
	// structure.
	pcn := conn.(*viaproxy.Conn)
	log.Printf("remote address is: %v", pcn.RemoteAddr())
	log.Printf("local address is: %v", pcn.LocalAddr())
	log.Printf("proxy address is: %v", pcn.ProxyAddr())
	pcn.Close()
}
```

In this case, `Accept` returns a generic [`net.Conn`](https://golang.org/pkg/net/#Conn) object. If you want to directly use a `Conn` object (which satisfies the `net.Conn` interface), you can use `AcceptFromProxy` instead:

```go
ln, err := viaproxy.Listen("tcp", *addr)
if err != nil {
	log.Fatal(err)
}

for {
	cn, err := ln.AcceptFromProxy()
	if err != nil {
		log.Println("ln.Accept():", err)
		continue
	}

	// The connection should be safe to be converted to a *viaproxy.Conn
	// structure.
	log.Printf("remote address is: %v", cn.RemoteAddr())
	log.Printf("local address is: %v", cn.LocalAddr())
	log.Printf("proxy address is: %v", cn.ProxyAddr())
	cn.Close()
}
```

## Caveats
* Only works with TCP connections.
* Both endpoints of the connection **must** be compatible with proxy protocol.

## License
See [LICENSE](LICENSE).
