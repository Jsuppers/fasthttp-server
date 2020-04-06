package aws

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"fasthttp-server/pipe"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	awsBucket       = "AWS_BUCKET"
	awsRegion       = "AWS_REGION"
	awsAccessKey    = "AWS_ACCESS_KEY"
	awsAccessSecret = "AWS_ACCESS_SECRET"
)

var logFatalf = log.Fatalf

type Streamer interface {
	Stream(reader pipe.Simple)
	Close()
}

type streamer struct {
	bucket, region, key, accessKey, accessSecret string
	running                                      sync.WaitGroup
}

func New(clientID int) Streamer {
	s := &streamer{}
	s.key = getKey(clientID)
	s.bucket = os.Getenv(awsBucket)
	s.region = os.Getenv(awsRegion)
	s.accessKey = os.Getenv(awsAccessKey)
	s.accessSecret = os.Getenv(awsAccessSecret)

	if s.bucket == "" || s.region == "" || s.accessKey == "" || s.accessSecret == "" {
		message := "Please ensure all environment variables are set, this includes:"
		logFatalf("%s\n%s\n%s\n%s\n%s\n", message, awsBucket, awsRegion, awsAccessKey, awsAccessSecret)
	}

	return s
}

func (s *streamer) Close() {
	s.running.Wait()
}

func (s *streamer) Stream(reader pipe.Simple) {
	awsConfig := &aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
	}

	sess := session.Must(session.NewSession(awsConfig))
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // upload in 5MB chunks (this is the minimum allowed)
		u.LeavePartsOnError = true
	})

	s.running.Add(1)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Body:   reader,
	})
	s.running.Done()
	if err != nil {
		log.Println("error when uploading", err)
		return
	}
	fmt.Println("finished Streaming")
}

func getKey(clientID int) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("/chat/%s/content_logs_%s_%d", date, date, clientID)
}
