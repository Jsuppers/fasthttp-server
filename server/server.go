package server

import (
	"fmt"
	"log"
	"net"

	"fasthttp-server/pipe"

	"github.com/valyala/fasthttp"
)

var dataPipe pipe.Simple

func Start(listener net.Listener, stream pipe.Simple) error {
	fmt.Println("Starting http server at address: ", listener.Addr())
	dataPipe = stream
	return fasthttp.Serve(listener, requestHandler)
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	_, err := dataPipe.Write(ctx.PostBody())
	if err != nil {
		log.Println("Error when reading request: ", err)
	}
}
