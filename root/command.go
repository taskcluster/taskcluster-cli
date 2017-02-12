package root

import "github.com/spf13/cobra"

var (
	// Command is the root of the command tree.
	Command = &cobra.Command{
		Use:   "taskcluster",
		Short: "TaskCluster cli client.",
		Long:  "Long description of TaskCluster",
	}
)
