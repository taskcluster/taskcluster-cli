package expandscope

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func getExpectedScopes(inputScopes []string, t *testing.T) string {
	params := &auth.SetOfScopes{
		Scopes: inputScopes,
	}

	rawPayload, err := json.Marshal(params)
	assert.NoError(t, err)
	payload := bytes.NewReader(rawPayload)

	req, err := http.NewRequest("GET", "https://auth.taskcluster.net/v1/scopes/expand", payload)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	var s auth.SetOfScopes
	err = json.Unmarshal(body, &s)
	assert.NoError(t, err)
	expectedScopes := strings.Join(s.Scopes, "\n")
	return expectedScopes
}

func TestExpandSingleScope(t *testing.T) {
	assert := assert.New(t)

	inputScopes := []string{"assume:project:taskcluster:tutorial"}
	actualScopes := expandScope(inputScopes)
	expectedScopes := getExpectedScopes(inputScopes, t)
	assert.Equal(expectedScopes, actualScopes, "Error occured: the expected and the actual scopes should be equal.")
}

func TestExpandMultipleScope(t *testing.T) {
	assert := assert.New(t)

	inputScopes := []string{"assume:project:taskcluster:tutorial", "assume:repo:hg.mozilla.org/try:*"}
	actualScopes := expandScope(inputScopes)
	expectedScopes := getExpectedScopes(inputScopes, t)
	assert.Equal(expectedScopes, actualScopes, "Error occured: the expected and the actual scopes should be equal.")
}
