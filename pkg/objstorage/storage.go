package objstorage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"os"
)

const (
	bucketName = "github-backup"
)

var client *storage.Client

func init() {
	ctx := context.Background()
	c, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte{})) // TODO auth
	check(err)
	client = c
}

func CopyToBucket(localSrcFile *os.File) {
	fmt.Printf("copying %s to bucket\n", localSrcFile.Name())
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
		check(err)
		if bytesRead > 0 {
			_, err := bucketWriter.Write(buf[:bytesRead])
			check(err)
		}
	}
	fmt.Printf("copied %s to bucket\n", localSrcFile.Name())
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
