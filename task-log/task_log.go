package taskLog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/taskcluster/taskcluster-cli/apis"
	"github.com/taskcluster/taskcluster-cli/extpoints"
)

type taskLog struct{}

func init() {
	extpoints.Register("task-log", taskLog{})
}

func (taskLog) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (taskLog) Summary() string {
	return "Outputs the logs for <taskID> as generated, and exits when completes."
}

func (taskLog) Usage() string {
	usage := "Usage: taskcluster task-log <taskID>\n"
	return usage
}

func (taskLog) Execute(context extpoints.Context) bool {
	taskID := context.Arguments["<taskID>"].(string)

	// Get route from services.go
	route := apis.Services["Queue"].BaseURL
	for _, entry := range apis.Services["Queue"].Entries {
		if entry.Name == "getLatestArtifact" {
			route += entry.Route
			break
		}
	}

	route = strings.Replace(route, "<taskId>", taskID, 1)
	route = strings.Replace(route, "<name>", "public/logs/live_backing.log", 1)

	body, _ := makeGetRequest(route)

	// Check if we got any errors, we are not expecting a json response.
	var raw map[string]interface{}
	json.Unmarshal(body, &raw)

	if len(raw) != 0 {
		// Error, most likely with the taskID
		if _, ok := raw["message"]; ok {
			panic(raw["message"])
		}
		return false
	}

	fmt.Println(string(body))
	return true

}

func makeGetRequest(path string) (b []byte, err error) {
	resp, err := http.Get(path)
	if err != nil {
		panic("Error making request to " + path)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}
