package server

import (
	"encoding/json"
	"fasthttp-server/aws"
	"fasthttp-server/pipe"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	pipeNew = pipe.New
	awsNew  = aws.New
)

type Server interface {
	Start() error
	Close()
	Wait()
}

type server struct {
	httpServer fasthttp.Server
	listener   net.Listener
	dataPipes  map[int]pipe.Simple
	streamers  map[int]aws.Streamer
	waitGroup  sync.WaitGroup
}

type Request struct {
	ClientID int `json:"client_id"`
}

func New(l net.Listener) Server {
	return &server{
		dataPipes:  map[int]pipe.Simple{},
		streamers:  map[int]aws.Streamer{},
		listener:   l,
		httpServer: fasthttp.Server{},
	}
}

func (s *server) Start() error {
	s.waitGroup.Add(1)
	fmt.Println("Starting http server at address: ", s.listener.Addr())
	s.httpServer.Handler = s.requestHandler
	return s.httpServer.Serve(s.listener)
}

func (s *server) requestHandler(ctx *fasthttp.RequestCtx) {
	var message Request
	err := json.Unmarshal(ctx.PostBody(), &message)
	if err != nil {
		fmt.Println("Error parsing request", err)
		return
	}

	_, exists := s.dataPipes[message.ClientID]
	if !exists {
		fmt.Println("Creating file upload for client ", message.ClientID)
		s.dataPipes[message.ClientID] = pipeNew()
		go func() {
			s.streamers[message.ClientID] = awsNew(message.ClientID)
			s.streamers[message.ClientID].Stream(s.dataPipes[message.ClientID])
		}()
	}

	_, err = s.dataPipes[message.ClientID].Write(ctx.PostBody())
	if err != nil {
		log.Println("Error when reading request: ", err)
	}
}

func (s *server) Close() {
	fmt.Println("Shutting down the server")
	err := s.httpServer.Shutdown()
	if err != nil {
		log.Println("Error when shutting down the server: ", err)
	}

	fmt.Println("Closing data pipes")
	for _, dp := range s.dataPipes {
		dp.Close()
	}

	fmt.Println("Closing s3 streamers")
	for _, stream := range s.streamers {
		stream.Close()
	}

	fmt.Println("Closed all streamers")
	s.waitGroup.Done()
}

func (s *server) Wait() {
	s.waitGroup.Wait()
}
