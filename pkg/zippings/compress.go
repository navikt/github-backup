package zippings

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CompressIt(src, compressedFilename string) error {
	archive, err := os.Create(compressedFilename)
	if err != nil {
		return err
	}
	zipWriter := zip.NewWriter(archive)
	defer archive.Close()
	defer zipWriter.Close()

	return filepath.Walk(src, func(file string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if shouldBeSkipped(compressedFilename, file) || fileInfo.IsDir() {
			fmt.Printf("skipping %s\n", file)
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		writer, err := zipWriter.Create(strings.TrimPrefix(file, src+"/"))
		if err != nil {
			return err
		}
		if _, err := io.Copy(writer, f); err != nil {
			return err
		}

		return nil
	})
}

func FilenameFor(reponame string) string {
	now := time.Now().Format("2006-01-02-15:04:05")
	withSlashes := fmt.Sprintf("ghbackup_%s_%s.zip", reponame, now)
	return strings.ReplaceAll(withSlashes, "/", "_")
}

func shouldBeSkipped(compressedFilename, filenameToTest string) bool {
	// do not re-compress our own file
	if strings.HasSuffix(compressedFilename, filenameToTest) {
		return true
	}
	return false
}