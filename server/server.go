package server

import (
	"encoding/json"
	"fasthttp-server/pipe"
	"fasthttp-server/storage"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/valyala/fasthttp"
)

const (
	partSize    = 5 * 1024 * 1024 // minimum allowed for s3 storage
	concurrency = 10
)

var (
	pipeNew  = pipe.NewGzipWriter
	s3New    = storage.NewS3Streamer
	azureNew = storage.NewAzureStreamer
)

type Server interface {
	Start() error
	Close()
	Wait()
}

type server struct {
	httpServer fasthttp.Server
	listener   net.Listener
	dataPipes  map[int]pipe.GzipWriter
	streamers  map[int]storage.MessageStreamer
	waitGroup  sync.WaitGroup
}

type Request struct {
	ClientID int `json:"client_id"`
}

func New(l net.Listener) Server {
	return &server{
		dataPipes:  map[int]pipe.GzipWriter{},
		streamers:  map[int]storage.MessageStreamer{},
		listener:   l,
		httpServer: fasthttp.Server{},
		waitGroup:  sync.WaitGroup{},
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
		s.dataPipes[message.ClientID] = pipeNew()
		go func() {
			s.streamers[message.ClientID] = getStreamer(message.ClientID)
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

	fmt.Println("Closing streamers")
	for _, stream := range s.streamers {
		stream.Wait()
	}

	fmt.Println("Closed all streamers")
	s.waitGroup.Done()
}

func (s *server) Wait() {
	s.waitGroup.Wait()
}

func getStreamer(clientID int) storage.MessageStreamer {
	if os.Getenv("STORAGE_TYPE") == "azure" {
		return azureNew(clientID, partSize, concurrency)
	}
	return s3New(clientID, partSize, concurrency)
}
