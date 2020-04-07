package server

import (
	"bufio"
	"fasthttp-server/pipe"
	"fasthttp-server/storage"
	"fmt"
	"reflect"
	"testing"
	"time"

	"fasthttp-server/mocks"

	"github.com/golang/mock/gomock"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

//go:generate mockgen -package=mocks -destination=./../mocks/pipe_mock.go fasthttp-server/pipe GzipWriter
//go:generate mockgen -package=mocks -destination=./../mocks/aws_mock.go fasthttp-server/storage S3
//go:generate mockgen -package=mocks -destination=./../mocks/net_mock.go net Listener

func TestNew(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"creates a new server with all properties"},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockListener := mocks.NewMockListener(mockCtrl)

			want := &server{
				dataPipes: map[int]pipe.GzipWriter{},
				streamers: map[int]storage.S3{},
				listener:  mockListener,
			}

			if got := New(mockListener); !reflect.DeepEqual(got, want) {
				t.Errorf("NewGzipWriter() = %v, want %v", got, want)
			}

			mockCtrl.Finish()
		})
	}
}

func Test_server_Start(t *testing.T) {
	tests := []struct {
		name    string
		request string
		setup   func(mockPipe *mocks.MockGzipWriter, MockS3 *mocks.MockS3)
	}{
		{"error parsing request", fmt.Sprintf("POST / HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", 1, "{"),
			func(mockPipe *mocks.MockGzipWriter, MockS3 *mocks.MockS3) {
				mockPipe.EXPECT().Write(gomock.Any()).Times(0)
				MockS3.EXPECT().Stream(gomock.Any()).Times(0)
			}},
		{"error writing to pipe", fmt.Sprintf("POST / HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", 2, "{}"),
			func(mockPipe *mocks.MockGzipWriter, MockS3 *mocks.MockS3) {
				mockPipe.EXPECT().Write(gomock.Any()).Times(1).Return(0, fmt.Errorf("error"))
				MockS3.EXPECT().Stream(gomock.Any()).Times(1)
			}},
		{"success", fmt.Sprintf("POST / HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", 2, "{}"),
			func(mockPipe *mocks.MockGzipWriter, MockS3 *mocks.MockS3) {
				mockPipe.EXPECT().Write(gomock.Any()).Times(1)
				MockS3.EXPECT().Stream(gomock.Any()).Times(1)
			}},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			ln := fasthttputil.NewInmemoryListener()
			mockCtrl := gomock.NewController(t)
			mockPipe := mocks.NewMockGzipWriter(mockCtrl)
			MockS3 := mocks.NewMockS3(mockCtrl)
			test.setup(mockPipe, MockS3)

			pipeNew = func() pipe.GzipWriter {
				return mockPipe
			}
			awsNew = func(int) storage.S3 {
				return MockS3
			}

			defer func() {
				awsNew = storage.NewS3Streamer
				pipeNew = pipe.NewGzipWriter
			}()

			s := &server{
				dataPipes: map[int]pipe.GzipWriter{},
				streamers: map[int]storage.S3{},
				listener:  ln,
			}

			// Start the server with an in memory listener
			serverCh := make(chan struct{})
			go func() {
				if err := s.Start(); err != nil {
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
				if _, err = c.Write([]byte(test.request)); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				var resp fasthttp.Response
				if err := resp.Read(br); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				if resp.StatusCode() != 200 {
					t.Errorf("unexpected status code: %d. Expecting %d", resp.StatusCode(), 200)
				}
				if err := c.Close(); err != nil {
					t.Errorf("unexpected error: %s", err)
				}
				// Give the server a little bit of time to transition the connection to the close state.
				time.Sleep(time.Millisecond * 100)
				close(clientCh)
			}()

			// close channels and listener
			<-clientCh
			_ = ln.Close()
			<-serverCh

			mockCtrl.Finish()
		})
	}
}

func Test_server_Close(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"successfully closes all resources"},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockPipe := mocks.NewMockGzipWriter(mockCtrl)
			MockS3 := mocks.NewMockS3(mockCtrl)

			dataPipes := map[int]pipe.GzipWriter{}
			dataPipes[0] = mockPipe
			mockPipe.EXPECT().Close().Times(1)

			streamers := map[int]storage.S3{}
			streamers[0] = MockS3
			MockS3.EXPECT().Wait().Times(1)

			s := &server{
				dataPipes: dataPipes,
				streamers: streamers,
			}

			s.waitGroup.Add(1)
			s.Close()

			mockCtrl.Finish()
		})
	}
}

func Test_server_Wait(t *testing.T) {
	s := &server{}
	s.waitGroup.Add(1)
	go s.waitGroup.Done()
	s.Wait()
	fmt.Println("Does not deadlock, that's good!")
}
