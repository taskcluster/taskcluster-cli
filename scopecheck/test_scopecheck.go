package scopecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestscopeCheck(t *testing.T) {

	assert := assert.New(t)

	myScopes := "['queue:ping']"
	testscope := "['queue:test']"

	assert.Equal("YES", checkscopes(myScopes, myScopes))
	assert.Equal("NO missing -queue:test", checkscopes(myScopes, testscope))

}
