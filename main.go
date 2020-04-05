package main

import (
	"fasthttp-server/server"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

const (
	network = "tcp"
	address = ":8080"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("Error creating listener: %s", err)
	}

	s := server.New(listener)

	go func() {
		sig := <-c
		fmt.Printf("Got %s signal. closing stream...\n", sig)
		s.Close()
		os.Exit(0)
	}()

	err = s.Start()
	if err != nil {
		log.Fatalf("Error starting fastHttp Server: %s", err)
	}

	s.Wait()
}
