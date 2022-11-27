package objstorage

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSlashesAreReplacedInFilenamed(t *testing.T) {
	filename := FilenameFor("navikt/whatever")
	assert.False(t, strings.Contains(filename, "/"))
}
