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
	for _, service := range args {
		fmt.Printf("%v status: %v\n", service, "running")
	}
	return nil
}
