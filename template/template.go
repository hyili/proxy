package template

import (
	"errors"
	"fmt"
	"io"
	"net"
	"proxy/template/structure"
)

// implicit proxy interface definition
type Proxy interface {
	ReadyToCommunicate() error
	clientToServer(int)
	serverToClient(int)
	CloseProxy()
}

// implement TCP proxy, according to proxy interface
type TcpProxy struct {
	NetProto, CliListenNetAddress, SrvNetAddress string

	cliListener              net.Listener             // interface
	cliConnPool, srvConnPool structure.ConnectionPool // interface

	connPairPool          structure.ConnectionPairPool
	cliErr, srvErr        error
	Done                  chan bool
	ClientToServerHandler func([]byte, int)
	ServerToClientHandler func([]byte, int)
}

func (proxy *TcpProxy) ReadyToCommunicate() error {
	var err error

	proxy.cliListener, err = net.Listen(proxy.NetProto, proxy.CliListenNetAddress)
	if err != nil {
		return err
	}

	if proxy.cliListener == nil {
		return errors.New("Listener not iniailized.")
	}

	structure.ConnectionPoolInit(&proxy.cliConnPool)
	structure.ConnectionPoolInit(&proxy.srvConnPool)
	structure.ConnectionPairPoolInit(&proxy.connPairPool)

	for {
		// Block here to wait
		fmt.Println(" [*] Listening for incoming client ...")
		cliConn, err := proxy.cliListener.Accept()
		if err != nil {
			return err
		}

		if cliConn != nil {
			fmt.Println(" [*] Got a client from " + cliConn.RemoteAddr().String() + "!")
			srvConn, err := net.Dial(proxy.NetProto, proxy.SrvNetAddress)
			if err != nil {
				return err
			}

			cid := proxy.cliConnPool.Add(cliConn)
			sid := proxy.srvConnPool.Add(srvConn)
			pid := proxy.connPairPool.Add(cid, sid)

			fmt.Println(" [*] Start communication.")
			if cliConn != nil && srvConn != nil {
				go proxy.clientToServer(pid)
				go proxy.serverToClient(pid)
			}
		}
	}

	return err
}

func (proxy *TcpProxy) clientToServer(pid int) {
	buffer := make([]byte, 1024)
	buflen := 0

	connPair := proxy.connPairPool.Get(pid)
	cliConn := proxy.cliConnPool.Get(connPair.Cid)
	srvConn := proxy.srvConnPool.Get(connPair.Sid)

	for {
		buflen, proxy.cliErr = cliConn.Read(buffer)

		if proxy.cliErr == io.EOF {
			fmt.Println(" [*] Client has terminated the connection.")
			connPair.Cancel()
			break
		} else if proxy.cliErr != nil {
			fmt.Println(" [*] Connection terminated.")
			return
		}

		proxy.ClientToServerHandler(buffer, buflen)
		srvConn.Write(buffer[0:buflen])

		select {
		case <-connPair.Ctx.Done():
			return
		default:
		}
	}

	// TODO: need to wait for serverToClient Recycle() before it can be accessed by new connection
	// TODO: must make sure that clientToSerevr() & serverToClient() will both return correctly
	// the only exception is that main() exit, the all goroutines will be terminated
	cliConn.Close()
	srvConn.Close()
	proxy.cliConnPool.Recycle(connPair.Cid)
	proxy.srvConnPool.Recycle(connPair.Sid)
	proxy.connPairPool.Recycle(pid)
}

func (proxy *TcpProxy) serverToClient(pid int) {
	buffer := make([]byte, 1024)
	buflen := 0

	connPair := proxy.connPairPool.Get(pid)
	cliConn := proxy.cliConnPool.Get(connPair.Cid)
	srvConn := proxy.srvConnPool.Get(connPair.Sid)

	for {
		buflen, proxy.srvErr = srvConn.Read(buffer)

		if proxy.srvErr == io.EOF {
			fmt.Println(" [*] Server has terminated the connection.")
			connPair.Cancel()
			break
		} else if proxy.srvErr != nil {
			fmt.Println(" [*] Connection terminated.")
			return
		}

		proxy.ServerToClientHandler(buffer, buflen)
		cliConn.Write(buffer[0:buflen])

		select {
		case <-connPair.Ctx.Done():
			return
		default:
		}
	}

	// TODO: need to wait for clientToServer Recycle() before it can be accessed by new connection
	// TODO: must make sure that clientToSerevr() & serverToClient() will both return correctly
	// the only exception is that main() exit, the all goroutines will be terminated
	cliConn.Close()
	srvConn.Close()
	proxy.cliConnPool.Recycle(connPair.Cid)
	proxy.srvConnPool.Recycle(connPair.Sid)
	proxy.connPairPool.Recycle(pid)
}

// TODO: Some problem here, would hang
func (proxy *TcpProxy) CloseProxy() {
	//	proxy.cliConnPool.DrainPool()
	//	proxy.srvConnPool.DrainPool()
	//	proxy.connPairPool.DrainPool()
}
