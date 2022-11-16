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
		err := CompressIt(basePath, tarGzPath, []string{})
		assert.NoError(t, err)
		fileStats, err := os.Stat(tarGzPath)
		assert.NoError(t, err)
		assert.True(t, fileStats.Size() > 0)
		assert.False(t, compressedFileContains(compressedFileName))
	})
}

func TestDenylistedFilesAreExcluded(t *testing.T) {
	withTestdata(func() {
		err := CompressIt(basePath, tarGzPath, []string{".git/"})
		assert.NoError(t, err)
		assert.True(t, compressedFileContains("file1"))
		assert.True(t, compressedFileContains("file2"))
		assert.False(t, compressedFileContains(".git/whatever"))
	})
}

func withTestdata(f func()) {
	createTestdata()
	f()
	cleanupTestdata()
}

func createTestdata() {
	os.MkdirAll(filepath.Join(basePath, ".git"), 0744)
	filenames := []string{"file1", "file2", ".git/whatever"}
	fileContents := []byte("hello")
	for _, f := range filenames {
		os.WriteFile(filepath.Join(basePath, f), fileContents, 0644)
	}
}

func cleanupTestdata() {
	os.RemoveAll(basePath)
}

func compressedFileContains(name string) bool {
	f, _ := os.Open(tarGzPath)
	defer f.Close()

	gzf, _ := gzip.NewReader(f)
	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if strings.Contains(header.Name, name) {
			return true
		}
	}
	return false
}
