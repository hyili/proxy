package template

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/hyili/proxy/template/structure"
	"io"
	"io/ioutil"
	"net"
	"strconv"
)

// implement TCP proxy, according to proxy interface
type TcpProxy struct {
	NetProto, CliListenNetAddress, SrvNetAddress string

	cliListener              net.Listener             // interface
	cliConnPool, srvConnPool structure.ConnectionPool // interface

	TlsOn           bool   // TLS
	certPem         []byte // TLS
	keyPem          []byte // TLS
	CertPemFilePath string // TLS
	KeyPemFilePath  string // TLS

	connPairPool          structure.ConnectionPairPool
	cliErr, srvErr        error
	Done                  chan bool
	ClientToServerHandler func([]byte, int, int, net.Addr, net.Addr)
	ServerToClientHandler func([]byte, int, int, net.Addr, net.Addr)
}

func (proxy *TcpProxy) ReadyToCommunicate() error {
	var err error
	var srvConn, cliConn net.Conn

	if proxy.TlsOn {
		fmt.Println(color.YellowString(" [*] ") + "Tls on ...")
		proxy.certPem, err = ioutil.ReadFile(proxy.CertPemFilePath)
		if err != nil {
			return err
		}

		proxy.keyPem, err = ioutil.ReadFile(proxy.KeyPemFilePath)
		if err != nil {
			return err
		}

		keyPair, err := tls.X509KeyPair(proxy.certPem, proxy.keyPem)
		if err != nil {
			return err
		}

		config := &tls.Config{Certificates: []tls.Certificate{keyPair}}
		proxy.cliListener, err = tls.Listen(proxy.NetProto, proxy.CliListenNetAddress, config)
	} else {
		fmt.Println(color.YellowString(" [*] ") + "Tls off ...")
		proxy.cliListener, err = net.Listen(proxy.NetProto, proxy.CliListenNetAddress)
	}

	if err != nil {
		return err
	}

	if proxy.cliListener == nil {
		return errors.New("Listener not iniailized.")
	}

	structure.ConnectionPoolInit(&proxy.cliConnPool)
	structure.ConnectionPoolInit(&proxy.srvConnPool)
	structure.ConnectionPairPoolInit(&proxy.connPairPool)

	defer proxy.CloseProxy()

	fmt.Println(color.YellowString(" [*] ") + "Listening on " + color.RedString(proxy.CliListenNetAddress) + " for incoming client ...")
	fmt.Println(color.YellowString(" [*] ") + "Aiming on server(" + color.MagentaString(proxy.SrvNetAddress) + ") ...")
	for {
		// Block here to wait
		cliConn, err = proxy.cliListener.Accept()
		if err != nil {
			return err
		}

		if cliConn != nil {
			fmt.Println(color.YellowString(" [*] ") + "Got a client(" + color.MagentaString(cliConn.RemoteAddr().String()) + ")!")
			if proxy.TlsOn {
				srvConn, err = tls.Dial(proxy.NetProto, proxy.SrvNetAddress, nil)
			} else {
				srvConn, err = net.Dial(proxy.NetProto, proxy.SrvNetAddress)
			}

			if err != nil {
				return err
			}

			cid := proxy.cliConnPool.Add(cliConn)
			sid := proxy.srvConnPool.Add(srvConn)
			pid := proxy.connPairPool.Add(cid, sid)

			fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Start communication.")
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
	cliAddr := cliConn.RemoteAddr()
	srvAddr := srvConn.RemoteAddr()

	for {
		buflen, proxy.cliErr = cliConn.Read(buffer)

		if proxy.cliErr == io.EOF {
			fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Client(" + color.MagentaString(cliAddr.String()) + ") has terminated the connection(" + color.MagentaString("pid="+strconv.Itoa(pid)+",cid="+strconv.Itoa(connPair.Cid)+",sid="+strconv.Itoa(connPair.Sid)) + ").")
			break
		} else if proxy.cliErr != nil {
			fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Client(" + color.MagentaString(cliAddr.String()) + ") connection(" + color.MagentaString("pid="+strconv.Itoa(pid)+",cid="+strconv.Itoa(connPair.Cid)+",sid="+strconv.Itoa(connPair.Sid)) + ") has been terminated.")
			fmt.Println(proxy.cliErr)
			break
		}

		proxy.ClientToServerHandler(buffer, buflen, pid, cliAddr, srvAddr)
		srvConn.Write(buffer[0:buflen])
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
	cliAddr := cliConn.RemoteAddr()
	srvAddr := srvConn.RemoteAddr()

	for {
		buflen, proxy.srvErr = srvConn.Read(buffer)

		if proxy.srvErr == io.EOF {
			fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Server(" + color.MagentaString(srvAddr.String()) + ") has terminated the connection(" + color.MagentaString("pid="+strconv.Itoa(pid)+",cid="+strconv.Itoa(connPair.Cid)+",sid="+strconv.Itoa(connPair.Sid)) + ").")
			break
		} else if proxy.srvErr != nil {
			fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Server(" + color.MagentaString(srvAddr.String()) + ") connection(" + color.MagentaString("pid="+strconv.Itoa(pid)+",cid="+strconv.Itoa(connPair.Cid)+",sid="+strconv.Itoa(connPair.Sid)) + ") has been terminated.")
			fmt.Println(proxy.srvErr)
			break
		}

		proxy.ServerToClientHandler(buffer, buflen, pid, cliAddr, srvAddr)
		cliConn.Write(buffer[0:buflen])
	}

	// TODO: need to wait for serverToClient Recycle() before it can be accessed by new connection
	// TODO: must make sure that clientToSerevr() & serverToClient() will both return correctly
	// the only exception is that main() exit, the all goroutines will be terminated
	cliConn.Close()
	srvConn.Close()
}

// TODO: Some problem here, would hang
func (proxy *TcpProxy) CloseProxy() {
	//	proxy.cliConnPool.DrainPool()
	//	proxy.srvConnPool.DrainPool()
	//	proxy.connPairPool.DrainPool()
}
