/*
Package viaproxy provides the ability to manage connections that properly understand the
Proxy Protocol defined by Willy Tarreau for HAProxy.

Regular net.Conn structures will return the "wrong" RemoteAddr when used behind a proxy:
the remote address informed will be the one of the proxy, not the one from the client
initiating the connection to the proxy. This package adds a wrapper for regular net.Conn
that checks for the existance of the proxy protocol line and if present, return a net.Conn
that reports the customer address when calling RemoteAddr. This wrapped connection can be
casted to *viaproxy.Conn to add access to an additional ProxyAddr method that will return
the proxy's address.

In order to use this extended connection type we can call Wrap on an existing connection:

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		conn, err = viaproxy.Wrap(conn)
		if err != nil {
			// handler error
		}

		cn := conn.(*viaproxy.Conn)
		log.Println(cn.ProxyAddr())
	}

Package viaproxy also provides a Listener struct that already returns viaproxy.Conn
connections when calling Accept:

	ln, err := viaproxy.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		cn := conn.(*viaproxy.Conn)
		log.Println(cn.ProxyAddr())
	}
*/
package viaproxy
