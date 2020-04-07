package storage

import (
	"bytes"
	"context"
	"fasthttp-server/mocks"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/golang/mock/gomock"
)

//go:generate mockgen -package=mocks -destination=./../mocks/azure_mock.go fasthttp-server/storage ContainerURL
//go:generate mockgen -package=mocks -destination=./../mocks/azblob_mock.go github.com/Azure/azure-storage-blob-go/azblob StorageError

func TestNewAzureStreamer(t *testing.T) {
	tests := []struct {
		name            string
		setup           func()
		want            *azure
		shouldCallFatal bool
	}{
		{"success", func() {
			os.Setenv(azureAccount, "azureAccount")
			os.Setenv(azureAccessKey, "azureAccessKey")
		}, &azure{
			blob:      getBlobName(0),
			account:   "azureAccount",
			accessKey: "azureAccessKey",
		}, false},
		{"should call error", func() {
		}, &azure{
			blob: getBlobName(0),
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
				os.Unsetenv(azureAccount)
				os.Unsetenv(azureAccessKey)
			}()

			if got := NewAzureStreamer(0, 0, 0); !reflect.DeepEqual(got, test.want) {
				t.Errorf("NewAzureStreamer() = %v, want %v", got, test.want)
			}

			if test.shouldCallFatal != fatal {
				t.Errorf("wanted %v but got %v", test.shouldCallFatal, fatal)
			}
		})
	}
}

func Test_azure_Wait(t *testing.T) {
	a := &azure{}
	a.running.Add(1)
	go a.running.Done()
	a.Wait()
	fmt.Println("Does not deadlock, that's good!")
}

func Test_getBlobName(t *testing.T) {
	date := time.Now().Format("2006-01-02")
	wanted := fmt.Sprintf("content-logs-%s-%d", date, 1)
	got := getBlobName(1)

	if got != wanted {
		t.Errorf("wanted %v but got %v", wanted, got)
	}
}

func Test_azure_Stream(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(calledUpload *bool, mockContainerURL *mocks.MockContainerURL, mockError *mocks.MockStorageError)
		shouldCallFatal bool
	}{
		{"should call the upload function", func(calledUpload *bool, mURL *mocks.MockContainerURL, e *mocks.MockStorageError) {
			mockFunctions(calledUpload, mURL)
			mURL.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			mURL.EXPECT().NewBlockBlobURL(gomock.Any()).Times(1)
		}, false},
		{"should call fatal if error is not ServiceCodeContainerAlreadyExists",
			func(calledUpload *bool, mURL *mocks.MockContainerURL, mError *mocks.MockStorageError) {
				mockFunctions(calledUpload, mURL)
				mURL.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil, mError)
				mURL.EXPECT().NewBlockBlobURL(gomock.Any()).Times(1)
				mError.EXPECT().ServiceCode().Times(1).Return(azblob.ServiceCodeSystemInUse)
			}, true},
		{"should call fatal credentials are invalid", func(calledUpload *bool, mURL *mocks.MockContainerURL, mError *mocks.MockStorageError) {
			mockFunctions(calledUpload, mURL)
			azblobNewSharedKeyCredential = func(accountName, accountKey string) (credential *azblob.SharedKeyCredential, err error) {
				return nil, fmt.Errorf("not valid")
			}
			mURL.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			mURL.EXPECT().NewBlockBlobURL(gomock.Any()).Times(1)
		}, true},
		{"should show error if upload fails", func(calledUpload *bool, mURL *mocks.MockContainerURL, mError *mocks.MockStorageError) {
			mockFunctions(calledUpload, mURL)
			azblobUploadStreamToBlockBlob =
				func(c context.Context, r io.Reader, b azblob.BlockBlobURL, o azblob.UploadStreamToBlockBlobOptions) (azblob.CommonResponse, error) {
					*calledUpload = true
					return nil, fmt.Errorf("error uploading")
				}
			mURL.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			mURL.EXPECT().NewBlockBlobURL(gomock.Any()).Times(1)
		}, false},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockContainerURL := mocks.NewMockContainerURL(mockCtrl)
			mockError := mocks.NewMockStorageError(mockCtrl)

			var buf bytes.Buffer
			var calledUpload bool

			s := &azure{
				blob:      getBlobName(0),
				account:   "azureAccount",
				accessKey: "azureAccessKey",
			}

			test.setup(&calledUpload, mockContainerURL, mockError)
			fatal := false
			logFatalf = func(format string, args ...interface{}) {
				fatal = true
			}
			defer func() {
				azblobNewSharedKeyCredential = azblob.NewSharedKeyCredential
				azblobUploadStreamToBlockBlob = azblob.UploadStreamToBlockBlob
				azblobNewContainerURL = NewContainerURL
				logFatalf = log.Fatalf
			}()

			s.Stream(&buf)

			if !calledUpload {
				t.Error("did not call upload function")
			}

			if test.shouldCallFatal != fatal {
				t.Errorf("wanted %v but got %v", test.shouldCallFatal, fatal)
			}

			mockCtrl.Finish()
		})
	}
}

func mockFunctions(calledUpload *bool, mURL *mocks.MockContainerURL) {
	azblobUploadStreamToBlockBlob =
		func(c context.Context, r io.Reader, b azblob.BlockBlobURL, o azblob.UploadStreamToBlockBlobOptions) (azblob.CommonResponse, error) {
			*calledUpload = true
			return nil, nil
		}
	azblobNewSharedKeyCredential = func(accountName, accountKey string) (credential *azblob.SharedKeyCredential, err error) {
		return nil, nil
	}
	azblobNewContainerURL = func(url *url.URL, p pipeline.Pipeline) ContainerURL {
		return mURL
	}
}

func TestNewContainerURL(t *testing.T) {
	u := url.URL{}
	want := azblob.NewContainerURL(u, nil)

	got := NewContainerURL(&u, nil)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("NewContainerURL() = %v, want %v", got, want)
	}
}
