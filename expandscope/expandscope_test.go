package expandscope

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func getExpectedScopes(inputScopes []string) (string, error) {
	params := &auth.SetOfScopes{
		Scopes: inputScopes,
	}
	var err error = nil

	rawPayload, err := json.Marshal(params)
	payload := bytes.NewReader(rawPayload)

	req, err := http.NewRequest("GET", "https://auth.taskcluster.net/v1/scopes/expand", payload)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var s auth.SetOfScopes
	err = json.Unmarshal(body, &s)

	expectedScopes := strings.Join(s.Scopes, "\n")
	return expectedScopes, err
}

func TestExpandSingleScope(t *testing.T) {
	assert := assert.New(t)

	inputScopes := []string{"assume:project:taskcluster:tutorial"}
	actualScopes := expandScope(inputScopes)
	expectedScopes, err := getExpectedScopes(inputScopes)
	if assert.NoError(err) {
		assert.Equal(expectedScopes, actualScopes, "Error occured: the expected and the actual scopes should be equal.")
	}
}

func TestExpandMultipleScope(t *testing.T) {
	assert := assert.New(t)

	inputScopes := []string{"assume:project:taskcluster:tutorial", "assume:repo:hg.mozilla.org/try:*"}
	actualScopes := expandScope(inputScopes)
	expectedScopes, err := getExpectedScopes(inputScopes)
	if assert.NoError(err) {
		assert.Equal(expectedScopes, actualScopes, "Error occured: the expected and the actual scopes should be equal.")
	}
}
