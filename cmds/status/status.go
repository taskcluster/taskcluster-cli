package task

import (
	"fmt"

	"github.com/taskcluster/taskcluster-cli/root"

	"github.com/spf13/cobra"
)

func init() {
	validArgs := []string{
		"queue",
		"auth",
		"awsprovisioner",
		"events",
		"index",
		"scheduler",
		"secrets",
	}
	use := "status"
	for _, validArg := range validArgs {
		use = use + " [" + validArg + "]"
	}
	statusCmd := &cobra.Command{
		Short: "taskcluster-cli status will query the current running status of taskcluster services",
		Long: `When called without arguments, taskcluster-clistatus will return the current running
status of all production taskcluster services.

By specifying one or more optional services as arguments, you can limit the
services included in the status report.`,
		PreRunE:   validateArgs,
		Use:       use,
		ValidArgs: validArgs,
		RunE:      status,
	}

	// Add the task subtree to the root.
	root.Command.AddCommand(statusCmd)
}

func validateArgs(cmd *cobra.Command, args []string) error {
outer:
	for _, arg := range args {
		for _, validArg := range cmd.ValidArgs {
			if arg == validArg {
				continue outer
			}
		}
		return fmt.Errorf("invalid argument(s) passed")
	}
	return nil
}

func status(cmd *cobra.Command, args []string) error {
	endpoints := pingEndpoints()

	for _, service := range args {
		fmt.Printf("%v status: %v\n", service, "running")
	}
	return nil
}

const jsonStream = `{
  "Auth": "http://references.taskcluster.net/auth/v1/api.json",
  "AuthEvents": "http://references.taskcluster.net/auth/v1/exchanges.json",
  "AwsProvisioner": "http://references.taskcluster.net/aws-provisioner/v1/api.json",
  "AwsProvisionerEvents": "http://references.taskcluster.net/aws-provisioner/v1/exchanges.json",
  "Github": "http://references.taskcluster.net/github/v1/api.json",
  "GithubEvents": "http://references.taskcluster.net/github/v1/exchanges.json",
  "Hooks": "http://references.taskcluster.net/hooks/v1/api.json",
  "Index": "http://references.taskcluster.net/index/v1/api.json",
  "Login": "http://references.taskcluster.net/login/v1/api.json",
  "Notify": "http://references.taskcluster.net/notify/v1/api.json",
  "Pulse": "http://references.taskcluster.net/pulse/v1/api.json",
  "PurgeCache": "http://references.taskcluster.net/purge-cache/v1/api.json",
  "PurgeCacheEvents": "http://references.taskcluster.net/purge-cache/v1/exchanges.json",
  "Queue": "http://references.taskcluster.net/queue/v1/api.json",
  "QueueEvents": "http://references.taskcluster.net/queue/v1/exchanges.json",
  "Scheduler": "http://references.taskcluster.net/scheduler/v1/api.json",
  "SchedulerEvents": "http://references.taskcluster.net/scheduler/v1/exchanges.json",
  "Secrets": "http://references.taskcluster.net/secrets/v1/api.json",
  "TreeherderEvents": "http://references.taskcluster.net/taskcluster-treeherder/v1/exchanges.json"
}`
