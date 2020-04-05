package server

import (
	"bufio"
	"testing"
	"time"

	"fasthttp-server/mocks"

	"github.com/golang/mock/gomock"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

//go:generate mockgen -package=mocks -destination=./../mocks/pipe_mock.go fasthttp-server/pipe Simple

func TestStart(t *testing.T) {
	StatusOK := 200
	ln := fasthttputil.NewInmemoryListener()
	mockCtrl := gomock.NewController(t)
	mockPipe := mocks.NewMockSimple(mockCtrl)
	mockPipe.EXPECT().Write(gomock.Any()).Times(1)

	// Start the server with an in memory listener
	serverCh := make(chan struct{})
	go func() {
		if err := Start(ln, mockPipe); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		close(serverCh)
	}()

	// Send the server a request and expect a response with status OK code
	clientCh := make(chan struct{})
	go func() {
		c, err := ln.Dial()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		br := bufio.NewReader(c)
		if _, err = c.Write([]byte("GET / HTTP/1.1\r\nHost: aa\r\n\r\n")); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		var resp fasthttp.Response
		if err := resp.Read(br); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if resp.StatusCode() != StatusOK {
			t.Errorf("unexpected status code: %d. Expecting %d", resp.StatusCode(), StatusOK)
		}
		if err := c.Close(); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		// Give the server a little bit of time to transition the connection to the close state.
		time.Sleep(time.Millisecond * 100)
		close(clientCh)
	}()

	// close channels and listener
	select {
	case <-clientCh:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	select {
	case <-serverCh:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	mockCtrl.Finish()
}
