package pipe

import (
	"compress/gzip"
	"io"
	"testing"
)

func TestNew(t *testing.T) {
	createdNewPipe := false
	createdGzipWriter := false

	ioPipe = func() (reader *io.PipeReader, writer *io.PipeWriter) {
		createdNewPipe = true
		return nil, nil
	}
	gzipNewWriter = func(w io.Writer) *gzip.Writer {
		createdGzipWriter = true
		return nil
	}

	defer func() {
		ioPipe = io.Pipe
		gzipNewWriter = gzip.NewWriter
	}()

	_ = New()

	if !createdNewPipe || !createdGzipWriter {
		t.Errorf("did not create all resources")
	}
}
