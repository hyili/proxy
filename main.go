package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"proxy/template"
	"syscall"
)

func ClientToServerHandler(buffer []byte, buflen int) {
	fmt.Println("Client said:")
	fmt.Println(string(buffer[:buflen]))
}

func ServerToClientHandler(buffer []byte, buflen int) {
	fmt.Println("Server said:")
	fmt.Println(string(buffer[:buflen]))
}

func main() {
	done := make(chan bool, 1)

	// create and register SIGINT SIGTERM SIGKILL to sigsChannel
	sigsChannel := make(chan os.Signal, 1)
	signal.Notify(sigsChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	P := template.TcpProxy{
		NetProto:            "tcp",
		CliListenNetAddress: "127.0.0.1:7077",
		SrvNetAddress:       "mail.itri.org.tw:25",
		Done:                done,
		ClientToServerHandler: ClientToServerHandler,
		ServerToClientHandler: ServerToClientHandler,
	}

	// defer
	defer P.CloseProxy()

	go func() {
		fmt.Println(" [*] ReadyToCommunicate ...")
		err := P.ReadyToCommunicate()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// signal goroutine
	go func() {
		sig := <-sigsChannel
		fmt.Println(" [*] " + sig.String())
		done <- true
	}()

	// hang in main program, wait for done channel's message
	<-done
}
