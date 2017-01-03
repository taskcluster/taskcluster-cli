package scopecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeCheck(t *testing.T) {

	assert := assert.New(t)

	myTestScope1 := "queue:ping"
	myTestScope3 := "queue:test"
	assert.Equal("YES", checkscopes(myTestScope1, myTestScope1))
	assert.Equal("NO missing -", checkscopes(myTestScope1, myTestScope3))

}
