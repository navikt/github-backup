package objstorage

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestStripPathFromFile(t *testing.T) {
	path := filepath.Join(os.TempDir(), "testfile")
	f, _ := os.Create(path)
	actual, _ := FilenameWithoutPath(f)
	assert.False(t, strings.Contains(actual, strconv.QuoteRune(os.PathSeparator)))
	_ = os.Remove(path)
}
