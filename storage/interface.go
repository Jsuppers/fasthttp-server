package storage

import (
	"io"
)

type MessageStreamer interface {
	Stream(reader io.Reader)
	Wait()
}
