package task

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestStatusString(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(getRunStatusString("only", ""), "only")
	assert.Equal(getRunStatusString("both", "here"), "both 'here'")
}
