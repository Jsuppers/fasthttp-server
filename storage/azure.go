package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	azureAccount   = "AZURE_STORAGE_ACCOUNT"
	azureAccessKey = "AZURE_STORAGE_ACCESS_KEY"
)

var (
	azblobUploadStreamToBlockBlob = azblob.UploadStreamToBlockBlob
	azblobNewSharedKeyCredential  = azblob.NewSharedKeyCredential
	azblobNewContainerURL         = NewContainerURL
)

type azure struct {
	blob, account, accessKey string
	bufferSize, maxBuffers   int
	running                  sync.WaitGroup
}

type ContainerURL interface {
	Create(ctx context.Context, metadata azblob.Metadata, publicAccessType azblob.PublicAccessType) (*azblob.ContainerCreateResponse, error)
	NewBlockBlobURL(blobName string) azblob.BlockBlobURL
}

func NewContainerURL(u *url.URL, p pipeline.Pipeline) ContainerURL {
	return azblob.NewContainerURL(*u, p)
}

func NewAzureStreamer(clientID, bufferSize, maxBuffers int) MessageStreamer {
	fmt.Println("Creating new Azure streamer for client ", clientID)
	s := &azure{}
	s.blob = getBlobName(clientID)
	s.account = os.Getenv(azureAccount)
	s.accessKey = os.Getenv(azureAccessKey)
	s.bufferSize = bufferSize
	s.maxBuffers = maxBuffers

	if s.account == "" || s.accessKey == "" {
		message := "Cannot create Azure streamer, ensure the following environment variables are set:"
		logFatalf("%s\n%s\n%s\n", message, azureAccount, azureAccessKey)
	}

	return s
}

func (a *azure) Stream(reader io.Reader) {
	credential, err := azblobNewSharedKeyCredential(a.account, a.accessKey)
	if err != nil {
		logFatalf("Invalid credentials with error: ", err)
	}

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", a.account, getContainerName()))
	containerURL := azblobNewContainerURL(URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	ctx := context.Background()
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok && serr.ServiceCode() != azblob.ServiceCodeContainerAlreadyExists {
			logFatalf("Error when creating container: ", err)
		}
	}

	a.running.Add(1)
	blobURL := containerURL.NewBlockBlobURL(a.blob)
	_, err = azblobUploadStreamToBlockBlob(ctx, reader, blobURL, azblob.UploadStreamToBlockBlobOptions{
		BufferSize: a.bufferSize,
		MaxBuffers: a.maxBuffers})
	a.running.Done()
	if err != nil {
		log.Println("Error when uploading", err)
	}
}

func getContainerName() string {
	return time.Now().Format("2006-01-02")
}

func getBlobName(clientID int) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("content-logs-%s-%d", date, clientID)
}

func (a *azure) Wait() {
	fmt.Println("Waiting for streaming to end for ", a.blob)
	a.running.Wait()
	fmt.Println("Finished Streaming to ", a.blob)
}
