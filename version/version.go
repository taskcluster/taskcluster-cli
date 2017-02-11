package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/taskcluster/taskcluster-cli/root"
)

var (
	// Command is the cobra command representing the version subtree.
	Command = &cobra.Command{
		Use:   "version",
		Short: "Prints the TaskCluster version.",
		Run:   printVersion,
	}

	// VersionNumber is a formatted string with the version information.
	VersionNumber = fmt.Sprintf("%d.%d.%d", 1, 0, 0)
)


func init() {
	root.Command.AddCommand(Command)
}

func printVersion(_ *cobra.Command, _ []string) {
	fmt.Printf("taskcluster (TaskCluster CLI) version %s\n", VersionNumber)
}
