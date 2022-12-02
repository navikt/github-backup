package objstorage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
)

var client *storage.Client

func init() {
	ctx := context.Background()
	c, err := storage.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	client = c
}

func CopyToBucket(localSrcFile *os.File, bucketName string) error {
	fmt.Printf("copying '%s' to bucket '%s'\n", localSrcFile.Name(), bucketName)
	ctx := context.Background()
	bucket := client.Bucket(bucketName)
	obj := bucket.Object(localSrcFile.Name())
	bucketWriter := obj.NewWriter(ctx)
	defer bucketWriter.Close()
	written, err := io.Copy(bucketWriter, localSrcFile)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d bytes to '%s'\n", written, bucketName)
	return nil
}
