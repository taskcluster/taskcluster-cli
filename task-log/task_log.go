package taskLog

import "github.com/taskcluster/taskcluster-cli/extpoints"

type taskLog struct{}

func init() {
	extpoints.Register("task-log", taskLog{})
}

func (taskLog) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (taskLog) Summary() string {
	return "Outputs the logs for <taskId> as generated, and exits when completes."
}

func (taskLog) Usage() string {
	usage := "Usage: taskcluster task-log <taskId>\n"
	return usage
}

func (taskLog) Execute(context extpoints.Context) bool {
	// TODO :
	// While task status is still running
	// print logs
	return true
}
