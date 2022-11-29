package zippings

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CompressIt(src, compressedFilename string, denyList []string) error {
	tarGzFile, err := os.OpenFile(compressedFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	defer tarGzFile.Close()
	if err != nil {
		return err
	}
	gzw := gzip.NewWriter(tarGzFile)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(src, func(file string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if shouldBeSkipped(compressedFilename, file, denyList) {
			return nil
		}
		var link string
		if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			if link, err = os.Readlink(file); err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(fileInfo, link)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(file)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			f, err := os.Open(file)
			defer f.Close()
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}

func shouldBeSkipped(compressedFilename, filenameToTest string, denylist []string) bool {
	// do not re-compress our own file
	if strings.HasSuffix(compressedFilename, filenameToTest) {
		return true
	}
	for _, path := range denylist {
		if strings.Contains(filenameToTest, path) {
			return true
		}
	}
	return false
}
