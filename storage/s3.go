package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

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
	message         = "Please ensure all environment variables are set, this includes:"
)

var (
	logFatalf            = log.Fatalf
	s3managerNewUploader = s3manager.NewUploader
)

type s3 struct {
	bucket, region, key, accessKey, accessSecret string
	running                                      sync.WaitGroup
}

func NewS3Streamer(clientID int) MessageStreamer {
	s := &s3{}
	s.key = getKey(clientID)
	s.bucket = os.Getenv(awsBucket)
	s.region = os.Getenv(awsRegion)
	s.accessKey = os.Getenv(awsAccessKey)
	s.accessSecret = os.Getenv(awsAccessSecret)

	if s.bucket == "" || s.region == "" || s.accessKey == "" || s.accessSecret == "" {
		logFatalf("%s\n%s\n%s\n%s\n%s\n", message, awsBucket, awsRegion, awsAccessKey, awsAccessSecret)
	}

	return s
}

func (s *s3) Stream(reader io.Reader) {
	awsConfig := &aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.NewStaticCredentials(s.accessKey, s.accessSecret, ""),
	}

	sess := session.Must(session.NewSession(awsConfig))
	uploader := s3managerNewUploader(sess, func(u *s3manager.Uploader) {
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
		log.Println("Error when uploading", err)
	}
}

func (s *s3) Wait() {
	fmt.Println("Waiting for streaming to end for ", s.key)
	s.running.Wait()
	fmt.Println("Finished Streaming to ", s.key)
}

func getKey(clientID int) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("/chat/%s/content_logs_%s_%d", date, date, clientID)
}
