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
	json.Unmarshal(ctx.PostBody(), &message)

	_, exists := s.dataPipes[message.ClientID]
	if !exists {
		fmt.Println("Creating file upload for client ", message.ClientID)
		s.dataPipes[message.ClientID] = pipeNew()
		go func() {
			s.streamers[message.ClientID] = awsNew(message.ClientID)
			s.streamers[message.ClientID].Stream(s.dataPipes[message.ClientID])
		}()
	}

	_, err := s.dataPipes[message.ClientID].Write(ctx.PostBody())
	if err != nil {
		log.Println("Error when reading request: ", err)
	}
}

func (s *server) Close() {

	fmt.Println("Shutting down the server")
	s.httpServer.Shutdown()

	fmt.Println("Closing data pipes")
	for _, pipe := range s.dataPipes {
		pipe.Close()
	}

	fmt.Println("Closing s3 streamers")
	for _, stream := range s.streamers {
		stream.Close()
	}
	s.waitGroup.Done()
}

func (s *server) Wait() {
	s.waitGroup.Wait()
}
