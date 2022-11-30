package zippings

import (
	"archive/zip"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var basePath = filepath.Join(os.TempDir(), "foo")
var compressedFileName = "compressed_stuff.zip"
var compressedFilePath = filepath.Join(basePath, compressedFileName)

func TestCompressedFileExcludesItself(t *testing.T) {
	withTestdata(func() {
		err := CompressIt(basePath, compressedFilePath)
		assert.NoError(t, err)
		fileStats, err := os.Stat(compressedFilePath)
		assert.NoError(t, err)
		assert.True(t, fileStats.Size() > 0)
		assert.False(t, compressedFileContains(compressedFileName))
	})
}

func TestSlashesAreReplacedInFilenamed(t *testing.T) {
	filename := FilenameFor("navikt/whatever")
	assert.False(t, strings.Contains(filename, "/"))
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
	archive, _ := zip.OpenReader(compressedFilePath)
	defer archive.Close()
	for _, f := range archive.File {
		fmt.Printf("'%s' contains '%s': %v\n", f.Name, name, strings.Contains(f.Name, name))
		return strings.Contains(f.Name, name)
	}
	return false
}
