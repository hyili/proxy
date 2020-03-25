package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

// implicit proxy interface definition
type Proxy interface {
	initializeProxy() error
	waitClient() error
	connServer() error
	showCommunication() error
	closeProxy() error
}

// implement TCP proxy, according to proxy interface
type TcpProxy struct {
	netProto, cliListenNetAddress, srvNetAddress string
	cliListener                                  net.Listener // interface
	cliConn, srvConn                             net.Conn     // interface
	cliErr, srvErr                               error
}

func (proxy TcpProxy) initializeProxy() error {
	proxy.cliListener, proxy.cliErr = net.Listen(proxy.netProto, proxy.cliListenNetAddress)

	return proxy.cliErr
}

func (proxy TcpProxy) waitClient() error {
	if proxy.cliListener == nil {
		// TODO
		proxy.cliErr = nil
	} else {
		proxy.cliConn, proxy.cliErr = proxy.cliListener.Accept()
	}

	return proxy.cliErr
}

func (proxy TcpProxy) connServer() error {
	proxy.srvConn, proxy.srvErr = net.Dial(proxy.netProto, proxy.srvNetAddress)

	return proxy.srvErr
}

func (proxy TcpProxy) clientToServer() []byte {
	buffer []byte
	proxy.cliErr := proxy.cliConn.Read(buffer)
	proxy.srvConn.
}

func (proxy TcpProxy) serverToClient() []byte {
}

func (proxy TcpProxy) showCommunication() error {

	return proxy.err
}

func main() {
	ln, err := net.Listen("tcp", ":7070")

	// Error Handler
	if err != nil {
		// stderr
		log.Fatalln(err)
	}

	// defer
	defer ln.Close()

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Fatalln(err)
			continue
		}
	}
}
