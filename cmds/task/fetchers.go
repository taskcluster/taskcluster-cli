package task

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/pflag"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

// allow overriding the base URL for testing
var queueBaseURL string

func makeQueue(credentials *tcclient.Credentials) *queue.Queue {
	q := queue.New(credentials)
	if queueBaseURL != "" {
		q.BaseURL = queueBaseURL
	}
	return q
}

// runStatus gets the status of run(s) of a given task.
func runStatus(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error {
	q := makeQueue(credentials)
	taskID := args[0]

	s, err := q.Status(taskID)
	if err != nil {
		return fmt.Errorf("could not get the status of the task %s: %v", taskID, err)
	}

	allRuns, _ := flagSet.GetBool("all-runs")
	runID, _ := flagSet.GetInt("run")

	if allRuns && runID != -1 {
		return fmt.Errorf("can't specify both all-runs and a specific run")
	}

	if allRuns {
		for _, r := range s.Status.Runs {
			fmt.Fprintf(out, "Run #%d: %s\n", r.RunID, getRunStatusString(r.State, r.ReasonResolved))
		}
		return nil
	}

	if runID >= len(s.Status.Runs) {
		return fmt.Errorf("there is no run #%v", runID)
	}
	if runID == -1 {
		runID = len(s.Status.Runs) - 1
	}

	fmt.Fprintln(out, getRunStatusString(s.Status.Runs[runID].State, s.Status.Runs[runID].ReasonResolved))
	return nil
}

// runName gets the name of a given task.
func runName(credentials *tcclient.Credentials, args []string, out io.Writer, _ *pflag.FlagSet) error {
	q := makeQueue(credentials)
	taskID := args[0]

	t, err := q.Task(taskID)
	if err != nil {
		return fmt.Errorf("could not get the task %s: %v", taskID, err)
	}

	fmt.Fprintln(out, t.Metadata.Name)
	return nil
}

// runGroup gets the groupID of a given task.
func runGroup(credentials *tcclient.Credentials, args []string, out io.Writer, _ *pflag.FlagSet) error {
	q := makeQueue(credentials)
	taskID := args[0]

	t, err := q.Task(taskID)
	if err != nil {
		return fmt.Errorf("could not get the task %s: %v", taskID, err)
	}

	fmt.Fprintln(out, t.TaskGroupID)
	return nil
}

// runArtifacts gets the name of the artificats for a given task and run.
func runArtifacts(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error {
	q := makeQueue(credentials)
	taskID := args[0]

	s, err := q.Status(taskID)
	if err != nil {
		return fmt.Errorf("could not get the status of the task %s: %v", taskID, err)
	}

	runID, _ := flagSet.GetInt("run")
	if runID >= len(s.Status.Runs) {
		return fmt.Errorf("there is no run #%v", runID)
	}
	if runID == -1 {
		runID = len(s.Status.Runs) - 1
	}

	buf := bytes.NewBufferString("")
	continuation := ""
	for {
		a, err := q.ListArtifacts(taskID, fmt.Sprint(runID), continuation, "")
		if err != nil {
			return fmt.Errorf("could not fetch artifacts for task %s run %v: %v", taskID, runID, err)
		}

		for _, ar := range a.Artifacts {
			fmt.Fprintf(buf, "%s\n", ar.Name)
		}

		continuation = a.ContinuationToken
		if continuation == "" {
			break
		}
	}

	buf.WriteTo(out)
	return nil
}

func runLog(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error {
	q := makeQueue(credentials)
	taskID := args[0]

	s, err := q.Status(taskID)
	if err != nil {
		return fmt.Errorf("could not get the status of the task %s: %v", taskID, err)
	}

	state := s.Status.State
	if state == "unscheduled" || state == "pending" {
		return fmt.Errorf("could not fetch the logs of task %s because it's in a %s state", taskID, state)
	}

	path := "https://queue.taskcluster.net/v1/task/" + taskID + "/artifacts/public/logs/live.log"

	resp, err := http.Get(path)
	if err != nil {
		return fmt.Errorf("Error making request to %v: %v", path, err)
	}
	defer resp.Body.Close()

	// Read line by line for live logs.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Fprintln(out, scanner.Text())
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Received unexpected response code %v", resp.StatusCode)
	}

	return nil
}
