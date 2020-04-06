package aws

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name            string
		setup           func()
		want            *streamer
		shouldCallFatal bool
	}{
		{"success", func() {
			os.Setenv(awsBucket, "awsBucket")
			os.Setenv(awsRegion, "awsRegion")
			os.Setenv(awsAccessKey, "awsAccessKey")
			os.Setenv(awsAccessSecret, "awsAccessSecret")
		}, &streamer{
			key:          getKey(0),
			bucket:       "awsBucket",
			region:       "awsRegion",
			accessKey:    "awsAccessKey",
			accessSecret: "awsAccessSecret",
		}, false},
		{"should call error", func() {
		}, &streamer{
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

			if got := New(0); !reflect.DeepEqual(got, test.want) {
				t.Errorf("New() = %v, want %v", got, test.want)
			}

			if test.shouldCallFatal != fatal {
				t.Errorf("wanted %v but got %v", test.shouldCallFatal, fatal)
			}
		})
	}
}
