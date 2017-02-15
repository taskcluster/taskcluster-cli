package taskLog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	route := "https://queue.taskcluster.net/v1/task/" + taskID + "/artifacts/public/logs/live.log"

	body, _ := makeGetRequest(route)

	// Check if we got any errors, we are not expecting a json response.
	var raw map[string]interface{}
	json.Unmarshal(body, &raw)

	if len(raw) != 0 {
		// Error, most likely with the taskID
		return false
	}

	return true

}

func makeGetRequest(path string) (b []byte, err error) {
	fmt.Println(path)
	resp, err := http.Get(path)
	if err != nil {
		panic("Error making request to " + path)
	}

	defer resp.Body.Close()

	// Read line by line for live logs.
	// This will also print the error message for failed requests.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}
