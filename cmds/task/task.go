// Package task  implements the task subcommands.
package task

import (
	"github.com/taskcluster/taskcluster-cli/cmds/root"

	"github.com/spf13/cobra"
)

var (
	// Command is the root of the task subtree.
	Command = &cobra.Command{
		Use:   "task",
		Short: "Provides task-related actions and commands.",
	}
	statusCmd = &cobra.Command{
		Use:   "status <taskId>",
		Short: "Get the status of a task.",
		RunE:  executeHelperE(runStatus),
	}
	artifactsCmd = &cobra.Command{
		Use:   "artifacts <taskId>",
		Short: "Get the name of the artifacts of a task.",
		RunE:  executeHelperE(runArtifacts),
	}
	awaitCmd = &cobra.Command{
		Use:   "await <taskId>",
		Short: "Watches the task and only returns on completion.",
		RunE:  executeHelperE(runAwait),
	}
)

func init() {
	statusCmd.Flags().BoolP("all-runs", "a", false, "Check all runs of the task.")
	statusCmd.Flags().IntP("run", "r", -1, "Specifies which run to consider.")

	artifactsCmd.Flags().IntP("run", "r", -1, "Specifies which run to consider.")

	awaitCmd.Flags().IntP("sleep", "s", 60, "Specifies how long to sleep between checks.")

	// Commands that fetch information
	Command.AddCommand(
		// status
		statusCmd,
		// name
		&cobra.Command{
			Use:   "name <taskId>",
			Short: "Get the name of a task.",
			RunE:  executeHelperE(runName),
		},
		// definition
		&cobra.Command{
			Use:   "def <taskId>",
			Short: "Get the full definition of a task.",
			RunE:  executeHelperE(runDef),
		},
		// group
		&cobra.Command{
			Use:   "group <taskId>",
			Short: "Get the taskGroupID of a task.",
			RunE:  executeHelperE(runGroup),
		},
		// artifacts
		artifactsCmd,
		// log
		&cobra.Command{
			Use:   "log <taskId>",
			Short: "Streams the log until completion.",
			RunE:  executeHelperE(runLog),
		},
		// await
		awaitCmd,
	)

	// Commands that take actions
	Command.AddCommand(
		// cancel
		&cobra.Command{
			Use:   "cancel <taskId>",
			Short: "Cancel a task.",
			RunE:  executeHelperE(runCancel),
		},
		// cancel
		&cobra.Command{
			Use:   "rerun <taskId>",
			Short: "Rerun a task.",
			RunE:  executeHelperE(runRerun),
		},
		// cancel
		&cobra.Command{
			Use:   "complete <taskId>",
			Short: "Complete the execution of a task.",
			RunE:  executeHelperE(runComplete),
		},
	)

	// Add the task subtree to the root.
	root.Command.AddCommand(Command)
}
