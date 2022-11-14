package zippings

import (
	"archive/tar"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var basePath = filepath.Join(os.TempDir(), "foo")
var compressedFileName = "compressed_stuff.tar.gz"
var tarGzPath = filepath.Join(basePath, compressedFileName)

func TestCompressedFileExcludesItself(t *testing.T) {
	withTestdata(func() {
		err := CompressIt(basePath, tarGzPath)
		assert.NoError(t, err)
		fileStats, err := os.Stat(tarGzPath)
		assert.NoError(t, err)
		assert.True(t, fileStats.Size() > 0)
		assert.False(t, compressedFileContainsItself())
	})
}

func withTestdata(f func()) {
	createTestdata()
	f()
	cleanupTestdata()
}

func createTestdata() {
	os.MkdirAll(basePath, 0744)
	filenames := []string{"file1", "file2"}
	fileContents := []byte("hello\ngo\n")
	for _, f := range filenames {
		os.WriteFile(filepath.Join(basePath, f), fileContents, 0644)
	}
}

func cleanupTestdata() {
	os.RemoveAll(basePath)
}

func compressedFileContainsItself() bool {
	f, _ := os.Open(tarGzPath)
	defer f.Close()

	gzf, _ := gzip.NewReader(f)
	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if strings.Contains(header.Name, compressedFileName) {
			return true
		}
	}
	return false
}
