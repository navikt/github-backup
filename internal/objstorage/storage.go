package objstorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

func CopyToBucket(gcsClient *storage.Client, localSrcFile *os.File, bucketName string, objBasePath string) error {
	srcFilename, err := FilenameWithoutPath(localSrcFile)
	if err != nil {
		return err
	}
	ctx := context.Background()
	bucket := gcsClient.Bucket(bucketName)
	objName := filepath.Join(objBasePath, srcFilename)
	fmt.Printf("copying '%s' to '%s' in bucket '%s'\n", srcFilename, objName, bucketName)
	obj := bucket.Object(objName)
	objWriter := obj.NewWriter(ctx)
	written, err := io.Copy(objWriter, localSrcFile)
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d bytes to '%s'\n", written, objName)
	return objWriter.Close()
}

func FilenameWithoutPath(f *os.File) (string, error) {
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	return info.Name(), nil
}
