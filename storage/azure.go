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

	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	azureAccount   = "AZURE_STORAGE_ACCOUNT"
	azureAccessKey = "AZURE_STORAGE_ACCESS_KEY"
)

type azure struct {
	blob, account, accessKey string
	running                  sync.WaitGroup
}

func NewAzureStreamer(clientID int) MessageStreamer {
	s := &azure{}
	s.blob = getBlobName(clientID)
	s.account = os.Getenv(azureAccount)
	s.accessKey = os.Getenv(azureAccessKey)

	if s.account == "" || s.accessKey == "" {
		logFatalf("%s\n%s\n%s\n%s\n%s\n", message, awsBucket, awsRegion, awsAccessKey, awsAccessSecret)
	}

	return s
}

func (a *azure) Stream(reader io.Reader) {
	credential, err := azblob.NewSharedKeyCredential(a.account, a.accessKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: ", err)
	}

	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", a.account, a.blob))
	containerURL := azblob.NewContainerURL(*URL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))

	ctx := context.Background() // This example uses a never-expiring context
	_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		log.Fatal("Error when creating container: ", err)
	}

	a.running.Add(1)
	blobURL := containerURL.NewBlockBlobURL(a.blob)
	_, err = azblob.UploadStreamToBlockBlob(ctx, reader, blobURL, azblob.UploadStreamToBlockBlobOptions{
		BufferSize: 4 * 1024 * 1024,
		MaxBuffers: 5})
	a.running.Done()
	if err != nil {
		log.Println("Error when uploading", err)
	}
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
