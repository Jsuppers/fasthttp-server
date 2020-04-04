package main

import (
	"fasthttp-server/server"
	"log"
	"net"
)

const (
	network = "tcp"
	address = "127.0.0.1:8080"
)

func main() {
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("Error creating listener: %s", err)
	}
	err = server.Start(listener)
	if err != nil {
		log.Fatalf("Error starting fastHttp Server: %s", err)
	}
}
