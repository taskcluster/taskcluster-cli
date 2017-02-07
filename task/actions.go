package task

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

type arguments map[string]interface{}

// SubCommand represents the function interface of the task subcommand.
type SubCommand func(credentials *tcclient.Credentials, args arguments) bool

func extractRunID(max int, param interface{}) (runID int, err error) {
	runID = max

	if param == nil {
		return
	}

	if str, ok := param.(string); ok {
		var id int
		if id, err = strconv.Atoi(str); err == nil {
			if id >= 0 && id < max {
				runID = id
			} else {
				err = fmt.Errorf("given runID is out of range: %v", id)
			}
		}
	} else {
		err = fmt.Errorf("runID is not a string: %v", str)
	}

	return
}

func getRunStatusString(state, resolved string) string {
	if resolved != "" {
		return fmt.Sprintf("%s '%s'", state, resolved)
	}

	return state
}

func (task) runStatus(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	s, err := q.Status(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get the status of the task %s: %v\n", taskID, err)
		return false
	}

	if args["--all-runs"].(bool) {
		for _, r := range s.Status.Runs {
			fmt.Printf("Run #%d: %s\n", r.RunID, getRunStatusString(r.State, r.ReasonResolved))
		}
		return true
	}

	runID, err := extractRunID(len(s.Status.Runs)-1, args["--run"])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid runID: %v\n", err)
		return false
	}

	fmt.Println(getRunStatusString(s.Status.Runs[runID].State, s.Status.Runs[runID].ReasonResolved))

	return true
}

func (task) runName(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	t, err := q.Task(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get the task %s: %v\n", taskID, err)
		return false
	}

	fmt.Println(t.Metadata.Name)

	return true
}

func (task) runGroup(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	t, err := q.Task(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get the task %s: %v\n", taskID, err)
		return false
	}

	fmt.Println(t.TaskGroupID)

	return true
}

func (task) runArtifacts(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	s, err := q.Status(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get the task %s: %v\n", taskID, err)
		return false
	}

	runID, err := extractRunID(len(s.Status.Runs)-1, args["--run"])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid runID: %v\n", err)
		return false
	}

	buf := bytes.NewBufferString("")
	continuation := ""
	for {
		a, err := q.ListArtifacts(taskID, fmt.Sprint(runID), continuation, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not fetch artifacts for task %s run %v: %v", taskID, runID, err)
			return false
		}

		for _, ar := range a.Artifacts {
			fmt.Fprintf(buf, "%s\n", ar.Name)
		}

		continuation = a.ContinuationToken
		if continuation == "" {
			break
		}
	}

	buf.WriteTo(os.Stdout)

	return true
}

func (task) runCancel(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	c, err := q.CancelTask(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not cancel the task %s: %v\n", taskID, err)
		return false
	}

	run := c.Status.Runs[len(c.Status.Runs)-1]
	fmt.Println(getRunStatusString(run.State, run.ReasonResolved))

	return true
}

func (task) runRerun(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	c, err := q.RerunTask(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not rerun the task %s: %v\n", taskID, err)
		return false
	}

	run := c.Status.Runs[len(c.Status.Runs)-1]
	fmt.Println(getRunStatusString(run.State, run.ReasonResolved))

	return true
}

func (task) runComplete(credentials *tcclient.Credentials, args arguments) bool {
	q := queue.New(credentials)
	taskID := args["<taskId>"].(string)

	s, err := q.Status(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get the status of the task %s: %v\n", taskID, err)
		return false
	}

	c, err := q.ClaimTask(taskID, fmt.Sprint(len(s.Status.Runs)-1), &queue.TaskClaimRequest{
		WorkerGroup: s.Status.WorkerType,
		WorkerID:    "taskcluster-cli",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not claim the task %s: %v\n", taskID, err)
		return false
	}

	wq := queue.New(&tcclient.Credentials{
		ClientID:    c.Credentials.ClientID,
		AccessToken: c.Credentials.AccessToken,
		Certificate: c.Credentials.Certificate,
	})
	r, err := wq.ReportCompleted(taskID, fmt.Sprint(c.RunID))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not complete the task %s: %v\n", taskID, err)
		return false
	}

	fmt.Println(getRunStatusString(r.Status.Runs[c.RunID].State, r.Status.Runs[c.RunID].ReasonResolved))

	return true
}
