package objstorage

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripPathFromFile(t *testing.T) {
	path := filepath.Join(os.TempDir(), "testfile")
	f, _ := os.Create(path)
	actual, _ := FilenameWithoutPath(f)
	assert.False(t, strings.Contains(actual, strconv.QuoteRune(os.PathSeparator)))
	_ = os.Remove(path)
}
