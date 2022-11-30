package objstorage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"os"
)

var client *storage.Client

func init() {
	ctx := context.Background()
	c, err := storage.NewClient(ctx, option.WithoutAuthentication())
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
	buf := make([]byte, 51200)
	for {
		bytesRead, err := localSrcFile.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if bytesRead > 0 {
			_, err := bucketWriter.Write(buf[:bytesRead])
			if err != nil {
				return err
			}
		}
	}
	fmt.Printf("copied %s to bucket\n", localSrcFile.Name())
	return nil
}
