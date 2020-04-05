package pipe

import (
	"fmt"
	"io"
)

type Simple interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close()
}

func New() Simple {
	r, w := io.Pipe()
	return &pipe{r, w, false}
}

type pipe struct {
	r      *io.PipeReader
	w      *io.PipeWriter
	finish bool
}

func (p *pipe) Read(b []byte) (int, error) {
	return p.r.Read(b)
}

func (p *pipe) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

func (p *pipe) Close() {
	if err := p.w.Close(); err != nil {
		fmt.Println("Got error when closing stream ", err)
	}
}
