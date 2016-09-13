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
	groupLogoBucket *storage.BucketHandle
)

type BucketType uint8

const (
	PostImage BucketType = iota
	GroupImage
	GroupLogo
)

func InitCloudStorage() {
	client, err := storage.NewClient(context.Background())
	postImageBucket = client.Bucket("zing-post-images");
	groupImageBucket = client.Bucket("zing-group-images");
	groupLogoBucket = client.Bucket("zing-group-logos");

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
	case GroupLogo:
		bucket = groupLogoBucket
		return bucket;
	}
	return bucket;
}

func getBucketName(bucketType BucketType) string {
	switch bucketType {
	case PostImage:
		return "zing-post-images"
	case GroupImage:
		bucket = groupImageBucket
		return "zing-group-images";
	case GroupLogo:
		bucket = groupLogoBucket
		return "zing-group-logos";
	}
	return "zing-post-images"
}

func SendItemToCloudStorage(bucketType BucketType, fileName string, fileType string, dec io.Reader) (string, error) {
	// Insert an object into a bucket.

	obj := getBucket(bucketType).Object(fileName);
	w := obj.NewWriter(context.Background());
	w.ContentType = "image/"
	w.ContentType += fileType
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	_, err := io.Copy(w, dec);
	if (err != nil) {
		slack.Send(slack.ErrorLevel, err.Error());
		return "", err
	}
	err = w.Close();
	if (err != nil) {
		slack.Send(slack.ErrorLevel, err.Error());
		return "", err
	}
	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, getBucketName(bucketType), fileName), err
}
