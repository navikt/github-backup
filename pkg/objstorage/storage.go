package objstorage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"os"
	"path/filepath"
	"time"
)

var client *storage.Client
var objBasePath = filepath.Join("/", time.Now().Format("2006/01/02"))

func init() {
	ctx := context.Background()
	c, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		panic(err)
	}
	client = c
}

func CopyToBucket(localSrcFile *os.File, bucketName string) error {
	srcFilename, err := FilenameWithoutPath(localSrcFile)
	if err != nil {
		return err
	}
	ctx := context.Background()
	bucket := client.Bucket(bucketName)
	objName := filepath.Join(objBasePath, srcFilename)
	fmt.Printf("copying '%s' to '%s' in bucket '%s'\n", srcFilename, objName, bucketName)
	obj := bucket.Object(objName)
	bucketWriter := obj.NewWriter(ctx)
	defer bucketWriter.Close()
	written, err := io.Copy(bucketWriter, localSrcFile)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d bytes to '%s'\n", written, objName)
	return nil
}

func FilenameWithoutPath(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	return info.Name(), nil
}
