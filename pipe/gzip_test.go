package pipe

import (
	"testing"
)

func TestRoundTrip(t *testing.T) {
	data := "h"
	pipe := NewGzipWriter()
	var buf = make([]byte, 64)
	go func() {
		n, err := pipe.Write([]byte(data))
		if err != nil {
			t.Errorf("write: %v", err)
		}
		if n != len(data) {
			t.Errorf("short write: %d != %d", n, len(data))
		}
		pipe.Close()
	}()
	n, err := pipe.Read(buf)
	if n == 0 {
		t.Errorf("error when reading from the stream: %v", err)
	}
	// TODO gunzip and check output
}
