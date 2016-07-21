package common

import "flag"
import (
	storage "google.golang.org/api/storage/v1"
	"golang.org/x/oauth2/google"
	"golang.org/x/net/context"
	"log"
	"fmt"
	"io"
)

const (
	scope = storage.DevstorageFullControlScope
)

var (
	projectID = flag.String("project", "fardo-854c8", "Your cloud project ID.")
	bucketName = flag.String("bucket", "post-images", "The name of an existing bucket within your project.")
	service *storage.Service
)

func InitCloudStorage() {
	client, err := google.DefaultClient(context.Background(), scope)
	if err != nil {
		log.Fatalf("Unable to get default client: %v", err)
	}
	service, err = storage.New(client)
	if err != nil {
		log.Fatalf("Unable to create storage service: %v", err)
	}
}

func SendItemToCloudStorage(itemType string, fileName string, file io.Reader) (*storage.Object, error) {
	// Insert an object into a bucket.
	object := &storage.Object{Name: fileName}

	object.Acl = [] *storage.ObjectAccessControl{{Entity : "allUsers", Role : "READER"} }
	res, err := service.Objects.Insert(*bucketName, object).Media(file).Do();
	if (err != nil) {
		return nil, err;
	}
	fmt.Printf("Created object %v at location %v\n\n", res.Name, res.SelfLink);
	return res, nil;
}

func GetObjectFromCloudStorage(fileName string) (*storage.Object, error) {
	// Get an object from a bucket.
	res, err := service.Objects.Get(*bucketName, fileName).Do();
	if (err != nil) {
		return nil, err
	}
	fmt.Printf("The media download link for %v/%v is %v.\n\n", *bucketName, res.Name, res.MediaLink)
	return res, nil
}
