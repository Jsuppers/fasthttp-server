package pipe

import (
	"compress/gzip"
	"fmt"
	"io"
)

var newLineBytes = []byte("\n")

type Simple interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close()
}

func New() Simple {
	r, w := io.Pipe()
	gw := gzip.NewWriter(w)
	return &pipe{r, w, gw}
}

type pipe struct {
	r  *io.PipeReader
	w  *io.PipeWriter
	gw *gzip.Writer
}

func (p *pipe) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func (p *pipe) Write(b []byte) (n int, err error) {
	n, err = p.gw.Write(b)
	if err != nil {
		return
	}
	return p.gw.Write(newLineBytes)
}

func (p *pipe) Close() {
	if err := p.w.Close(); err != nil {
		fmt.Println("Got error when closing stream ", err)
	}
}
