package task

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/pflag"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

// Executor represents the function interface of the task subcommand.
type Executor func(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error

// getRunStatusString takes the state and resolved strings and crafts a printable summary string.
func getRunStatusString(state, resolved string) string {
	if resolved != "" {
		return fmt.Sprintf("%s '%s'", state, resolved)
	}

	return state
}

// runStatus gets the status of run(s) of a given task.
func runStatus(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error {
	q := queue.New(credentials)
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
	q := queue.New(credentials)
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
	q := queue.New(credentials)
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
	q := queue.New(credentials)
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

// runCancel cancels the runs of a given task.
func runCancel(credentials *tcclient.Credentials, args []string, out io.Writer, _ *pflag.FlagSet) error {
	q := queue.New(credentials)
	taskID := args[0]

	c, err := q.CancelTask(taskID)
	if err != nil {
		return fmt.Errorf("could not cancel the task %s: %v", taskID, err)
	}

	run := c.Status.Runs[len(c.Status.Runs)-1]
	fmt.Fprintln(out, getRunStatusString(run.State, run.ReasonResolved))
	return nil
}

// runRerun re-runs a given task.
func runRerun(credentials *tcclient.Credentials, args []string, out io.Writer, _ *pflag.FlagSet) error {
	q := queue.New(credentials)
	taskID := args[0]

	c, err := q.RerunTask(taskID)
	if err != nil {
		return fmt.Errorf("could not rerun the task %s: %v", taskID, err)
	}

	run := c.Status.Runs[len(c.Status.Runs)-1]
	fmt.Fprintln(out, getRunStatusString(run.State, run.ReasonResolved))
	return nil
}

// runComplete completes a given task.
func runComplete(credentials *tcclient.Credentials, args []string, out io.Writer, _ *pflag.FlagSet) error {
	q := queue.New(credentials)
	taskID := args[0]

	s, err := q.Status(taskID)
	if err != nil {
		return fmt.Errorf("could not get the status of the task %s: %v", taskID, err)
	}

	c, err := q.ClaimTask(taskID, fmt.Sprint(len(s.Status.Runs)-1), &queue.TaskClaimRequest{
		WorkerGroup: s.Status.WorkerType,
		WorkerID:    "taskcluster-cli",
	})
	if err != nil {
		return fmt.Errorf("could not claim the task %s: %v", taskID, err)
	}

	wq := queue.New(&tcclient.Credentials{
		ClientID:    c.Credentials.ClientID,
		AccessToken: c.Credentials.AccessToken,
		Certificate: c.Credentials.Certificate,
	})
	r, err := wq.ReportCompleted(taskID, fmt.Sprint(c.RunID))
	if err != nil {
		return fmt.Errorf("could not complete the task %s: %v", taskID, err)
	}

	fmt.Fprintln(out, getRunStatusString(r.Status.Runs[c.RunID].State, r.Status.Runs[c.RunID].ReasonResolved))
	return nil
}
