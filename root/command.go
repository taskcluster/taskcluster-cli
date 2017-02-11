package root

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Command is the root of the command tree.
	Command *cobra.Command
)

func init() {
	Command = &cobra.Command{
		Use:   "taskcluster",
		Short: "Short description of TaskCluster",
		Long:  "Long description of TaskCluster",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Hello tc peeps\n")
		},
	}
}
