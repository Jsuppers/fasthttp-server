package main

import (
	"fasthttp-server/server"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	network = "tcp"
	address = ":8080"
)

func main() {
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("Error creating listener: %s", err)
	}

	s := server.New(listener)
	go closeGracefully(s)

	err = s.Start()
	if err != nil {
		log.Fatalf("Error starting fastHttp Server: %s", err)
	}

	s.Wait()
}

func closeGracefully(s server.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGQUIT)
	signal.Notify(c, syscall.SIGTERM)
	sig := <-c
	fmt.Printf("Got %s signal. closing stream...\n", sig)
	s.Close()
	os.Exit(0)
}
