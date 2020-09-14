package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net"
	"os"
	"os/signal"
	"proxy/template"
	"strconv"
	"syscall"
)

func ClientToServerHandler(buffer []byte, buflen int, pid int, cliAddr net.Addr, srvAddr net.Addr) {
	fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Client(" + color.MagentaString(cliAddr.String()) + ") said:")
	fmt.Println(string(buffer[:buflen]))
}

func ServerToClientHandler(buffer []byte, buflen int, pid int, cliAddr net.Addr, srvAddr net.Addr) {
	fmt.Println(color.YellowString(" ["+strconv.Itoa(pid)+"] ") + "Server(" + color.MagentaString(srvAddr.String()) + ") said:")
	fmt.Println(string(buffer[:buflen]))
}

func main() {
	done := make(chan bool, 1)

	// command-line arguments
	CliListenNetAddress := flag.String("listen", "127.0.0.1:7077", "Proxy will listen on this network address.\ndefault: 127.0.0.1:7077")
	SrvNetAddress := flag.String("connect", "hyili.idv.tw:443", "Proxy will connect to this network address.\ndefault: hyili.idv.tw:443")
	TlsOn := flag.Bool("tls", true, "Proxy will use tls or not.\ndefault: true")
	CertPemFilePath := flag.String("cert", "./cert.pem", "Certificate file path that will be used by Proxy to listen.\ndefault: ./cert.pem")
	KeyPemFilePath := flag.String("key", "./key.pem", "Key file path that will be used by Proxy to listen.\ndefault: ./key.pem")
	flag.Parse()

	// create and register SIGINT SIGTERM SIGKILL to sigsChannel
	sigsChannel := make(chan os.Signal, 1)
	signal.Notify(sigsChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	P := template.TcpProxy{
		NetProto:            "tcp",
		CliListenNetAddress: *CliListenNetAddress,
		SrvNetAddress:       *SrvNetAddress,
		TlsOn:               *TlsOn,
		CertPemFilePath:     *CertPemFilePath,
		KeyPemFilePath:      *KeyPemFilePath,
		Done:                done,
		ClientToServerHandler: ClientToServerHandler,
		ServerToClientHandler: ServerToClientHandler,
	}

	// defer
	defer P.CloseProxy()

	go func() {
		fmt.Println(color.YellowString(" [*] ") + "ReadyToCommunicate ...")
		err := P.ReadyToCommunicate()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// signal goroutine
	go func() {
		sig := <-sigsChannel
		fmt.Println(color.YellowString(" [*] ") + sig.String())
		done <- true
	}()

	// hang in main program, wait for done channel's message
	<-done
}
