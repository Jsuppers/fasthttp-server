package main

import (
	"fasthttp-server/aws"
	"fasthttp-server/pipe"
	"fasthttp-server/server"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

const (
	network = "tcp"
	address = "127.0.0.1:8080"
)

func main() {
	dataPipe := pipe.New()
	var streamer aws.Streamer

	go func() {
		streamer = aws.New()
		streamer.Stream(dataPipe)
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("Error creating listener: %s", err)
	}

	go func() {
		sig := <-c
		fmt.Printf("Got %s signal. closing stream...\n", sig)
		dataPipe.Close()
		streamer.Close()
		os.Exit(0)
	}()

	err = server.Start(listener, dataPipe)
	if err != nil {
		log.Fatalf("Error starting fastHttp Server: %s", err)
	}
}
