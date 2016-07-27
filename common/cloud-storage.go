package common

import (
	"google.golang.org/cloud/storage"
	"golang.org/x/net/context"
	"log"
	"io"
	"fmt"
	"github.com/mymachine8/fardo-api/slack"
)

var (
	bucket *storage.BucketHandle
	postImageBucket *storage.BucketHandle
	groupImageBucket *storage.BucketHandle
)

type BucketType uint8

const (
	PostImage BucketType = iota
	GroupImage
)

func InitCloudStorage() {
	client, err := storage.NewClient(context.Background())
	postImageBucket = client.Bucket("post-images");
	groupImageBucket = client.Bucket("group-images");

	if err != nil {
		slack.Send(slack.ErrorLevel, "New Client Creation Error: " + err.Error());
		log.Print(err)
	}
}

func getBucket(bucketType BucketType) *storage.BucketHandle {
	switch bucketType {
	case PostImage:
		bucket = postImageBucket
		return bucket;
	case GroupImage:
		bucket = groupImageBucket
		return bucket;
	}
	return bucket;
}

func getBucketName(bucketType BucketType) string {
	switch bucketType {
	case PostImage:
		return "post-images"
	case GroupImage:
		bucket = groupImageBucket
		return "group-images";
	}
	return "post-images"
}

func SendItemToCloudStorage(bucketType BucketType, fileName string,dec io.Reader) (string, error) {
	// Insert an object into a bucket.

	obj := getBucket(bucketType).Object(fileName);
	w := obj.NewWriter(context.Background());
	w.ContentType = "image/jpeg"
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	_,err := io.Copy(w, dec);
	if (err != nil) {
		slack.Send(slack.ErrorLevel, err.Error());
		return "", err
	}
	err = w.Close();
	if(err != nil) {
		slack.Send(slack.ErrorLevel, err.Error());
		return "", err
	}
	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, getBucketName(bucketType), fileName), err
}

/*func GetObjectFromCloudStorage(fileName string) (*storage.Object, error) {
	// Get an object from a bucket.
	res, err := service.Objects.Get(*bucketName, fileName).Do();
	if (err != nil) {
		return nil, err
	}
	fmt.Printf("The media download link for %v/%v is %v.\n\n", *bucketName, res.Name, res.MediaLink)
	return res, nil
}*/
