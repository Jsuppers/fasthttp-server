package storage

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func TestNewS3Streamer(t *testing.T) {
	tests := []struct {
		name            string
		setup           func()
		want            *s3
		shouldCallFatal bool
	}{
		{"success", func() {
			os.Setenv(awsBucket, "awsBucket")
			os.Setenv(awsRegion, "awsRegion")
			os.Setenv(awsAccessKey, "awsAccessKey")
			os.Setenv(awsAccessSecret, "awsAccessSecret")
		}, &s3{
			key:          getKey(0),
			bucket:       "awsBucket",
			region:       "awsRegion",
			accessKey:    "awsAccessKey",
			accessSecret: "awsAccessSecret",
		}, false},
		{"should call error", func() {
		}, &s3{
			key: getKey(0),
		}, true},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			test.setup()

			fatal := false
			logFatalf = func(format string, args ...interface{}) {
				fatal = true
			}
			defer func() {
				logFatalf = log.Fatalf
				os.Unsetenv(awsBucket)
				os.Unsetenv(awsRegion)
				os.Unsetenv(awsAccessKey)
				os.Unsetenv(awsAccessSecret)
			}()

			if got := NewS3Streamer(0); !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewS3Streamer() = %v, want %v", got, test.want)
			}

			if test.shouldCallFatal != fatal {
				t.Errorf("wanted %v but got %v", test.shouldCallFatal, fatal)
			}
		})
	}
}

func Test_server_Wait(t *testing.T) {
	s := &s3{}
	s.running.Add(1)
	go s.running.Done()
	s.Wait()
	fmt.Println("Does not deadlock, that's good!")
}

func Test_getKey(t *testing.T) {
	date := time.Now().Format("2006-01-02")
	wanted := fmt.Sprintf("/chat/%s/content_logs_%s_%d", date, date, 1)
	got := getKey(1)

	if got != wanted {
		t.Errorf("wanted %v but got %v", wanted, got)
	}
}

func Test_s3_Stream(t *testing.T) {
	tests := []struct {
		name  string
		setup func(calledUpload *bool)
	}{
		{"should call the upload function", func(calledUpload *bool) {
			s3managerNewUploader = func(c client.ConfigProvider, options ...func(*s3manager.Uploader)) *s3manager.Uploader {
				*calledUpload = true
				return s3manager.NewUploader(c, options...)
			}
		}},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			var calledUpload bool

			s := &s3{
				key:          getKey(0),
				bucket:       "awsBucket",
				region:       "awsRegion",
				accessKey:    "awsAccessKey",
				accessSecret: "awsAccessSecret",
			}

			test.setup(&calledUpload)
			defer func() {
				s3managerNewUploader = s3manager.NewUploader
			}()

			s.Stream(&buf)

			if !calledUpload {
				t.Error("did not call upload function")
			}
		})
	}
}
