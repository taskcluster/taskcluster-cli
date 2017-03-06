package task

import (
	"github.com/taskcluster/taskcluster-cli/root"

	"github.com/spf13/cobra"
)

var (
	// Command is the root of the task subtree.
	Command = &cobra.Command{
		Use:   "status",
		Short: "Shows the live status of taskcluster services.",
	}
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status <taskId>",
		Short: "Get the status of a service.",
		RunE:  status,
	}

	// Commands that fetch information
	Command.AddCommand(
		// status
		statusCmd,
	)

	// Add the task subtree to the root.
	root.Command.AddCommand(Command)
}

func status(*cobra.Command, []string) error {
	return nil
}
