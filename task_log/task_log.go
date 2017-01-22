package task_log

import "github.com/taskcluster/taskcluster-cli/extpoints"

type task_log struct{}

func init() {
	extpoints.Register("task-log", task_log{})
}

func (task_log) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (task_log) Summary() string {
	return "Outputs the logs for <taskId> as generated, and exits when completes."
}

func (task_log) Usage() string {
	usage := "Usage: taskcluster task-log <taskId>\n"
	return usage
}

func (task_log) Execute(context extpoints.Context) bool {
	// TODO :
	// While task status is still running
	// print logs
	return true
}
