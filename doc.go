/*
Package viaproxy provides the ability to manage connections that properly understand the
Proxy Protocol defined by Willy Tarreau for HAProxy.

Regular net.Conn structures will return the "wrong" RemoteAddr when used behind a proxy:
the remote address informed will be the one of the proxy, not the one from the client
initiating the connection to the proxy. This package adds a wrapper for regular net.Conn
that checks for the existence of the proxy protocol line and if present, return a net.Conn
that reports the customer address when calling RemoteAddr. This wrapped connection can be
casted to *viaproxy.Conn to add access to an additional ProxyAddr method that will return
the proxy's address.

In order to use this extended connection type we can call Wrap on an existing connection:

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

Package viaproxy also provides a Listener struct that already returns viaproxy.Conn
connections when calling AcceptFromProxy:

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

		log.Printf("remote address is: %v", cn.RemoteAddr())
		log.Printf("local address is: %v", cn.LocalAddr())
		log.Printf("proxy address is: %v", cn.ProxyAddr())
		cn.Close()
	}

The Accept method in the Listener struct returns a generic net.Conn
which can safely be casted to a viaproxy.Conn:

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

Using viaproxy.Conn objects whenever a net.Conn is expected should be
safe in all cases. If you encounter an issue please send a bug report
to https://github.com/inkel/viaproxy/issues
*/
package viaproxy
