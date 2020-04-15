package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"proxy/template"
	"syscall"
)

func main() {
	done := make(chan bool, 1)

	// create and register SIGINT SIGTERM SIGKILL to sigsChannel
	sigsChannel := make(chan os.Signal, 1)
	signal.Notify(sigsChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	P := template.TcpProxy{
		NetProto:            "tcp",
		CliListenNetAddress: "127.0.0.1:7070",
		SrvNetAddress:       "mail.itri.org.tw:25",
		Done:                done,
	}

	// defer
	defer P.CloseProxy()

	fmt.Println(" [*] ReadyToCommunicate ...")
	err := P.ReadyToCommunicate()
	if err != nil {
		log.Fatalln(err)
	}

	// signal goroutine
	go func() {
		sig := <-sigsChannel
		fmt.Println(" [*] " + sig.String())
		done <- true
	}()

	// hang in main program, wait for done channel's message
	<-done
}
