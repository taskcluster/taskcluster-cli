package group

import (
	"github.com/spf13/cobra"
	"github.com/taskcluster/taskcluster-cli/root"
)

var (
	// Command is the root of the group subtree.
	Command = &cobra.Command{
		Use:   "group",
		Short: "Provides group-related actions and commands.",
	}
)

func init() {
	cancelCmd := &cobra.Command{
		Use:   "cancel <groupId>",
		Short: "Cancel a whole group by groupId.",
		RunE:  executeHelperE(runCancel),
	}
	Command.AddCommand(cancelCmd)

	root.Command.AddCommand(Command)
}
