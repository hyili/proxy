package template

import (
	"errors"
	"fmt"
	"io"
	"net"
)

// implicit proxy interface definition
type Proxy interface {
	ReadyToCommunicate() error
	clientToServer()
	serverToClient()
	CloseProxy()
}

// implement TCP proxy, according to proxy interface
type TcpProxy struct {
	NetProto, CliListenNetAddress, SrvNetAddress string

	cliListener      net.Listener // interface
	cliConn, srvConn net.Conn     // interface
	cliErr, srvErr   error
	Done             chan bool
}

func clientToServerHandler(buffer []byte, buflen int) {
	// TODO: do something here
}

func serverToClientHandler(buffer []byte, buflen int) {
	// TODO: do something here
}

func (proxy TcpProxy) ReadyToCommunicate() error {
	var err error

	proxy.cliListener, err = net.Listen(proxy.NetProto, proxy.CliListenNetAddress)
	if err != nil {
		return err
	}

	if proxy.cliListener == nil {
		return errors.New("Listener not iniailized.")
	} else {
		// Block here to wait
		fmt.Println(" [*] Listening for incoming client ...")
		proxy.cliConn, err = proxy.cliListener.Accept()
		if err != nil {
			return err
		}
	}

	if proxy.cliConn != nil {
		fmt.Println(" [*] Got a client from " + proxy.cliConn.RemoteAddr().String() + "!")
		proxy.srvConn, err = net.Dial(proxy.NetProto, proxy.SrvNetAddress)

		fmt.Println(" [*] Start communication.")
		if proxy.cliConn != nil && proxy.srvConn != nil {
			go proxy.clientToServer()
			go proxy.serverToClient()
		}
	}

	return err
}

func (proxy TcpProxy) clientToServer() {
	buffer := make([]byte, 1024)
	buflen := 0

	for {
		buflen, proxy.cliErr = proxy.cliConn.Read(buffer)

		if proxy.cliErr == io.EOF {
			fmt.Println(" [*] Client has terminated the connection.")
			break
		}

		clientToServerHandler(buffer, buflen)

		proxy.srvConn.Write(buffer)
	}

	proxy.Done <- true
}

func (proxy TcpProxy) serverToClient() {
	buffer := make([]byte, 1024)
	buflen := 0

	for {
		buflen, proxy.srvErr = proxy.srvConn.Read(buffer)

		if proxy.cliErr == io.EOF {
			fmt.Println(" [*] Server has terminated the connection.")
			break
		}

		serverToClientHandler(buffer, buflen)

		proxy.cliConn.Write(buffer)
	}

	proxy.Done <- true
}

func (proxy TcpProxy) CloseProxy() {
	if proxy.srvConn != nil {
		proxy.cliConn.Close()
	}

	if proxy.cliConn != nil {
		proxy.srvConn.Close()
	}
}
