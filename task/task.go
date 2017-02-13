package task

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taskcluster/taskcluster-cli/config"
	"github.com/taskcluster/taskcluster-cli/root"

	tcclient "github.com/taskcluster/taskcluster-client-go"
)

var (
	// Command is the root of the
	Command = &cobra.Command{
		Use:   "task",
		Short: "Provides task-related actions and commands.",
	}
)

func init() {
	statusCmd := &cobra.Command{
		Use:   "status <taskId>",
		Short: "Get the status of a task.",
		RunE:  executeHelperE(runStatus),
	}
	statusCmd.Flags().BoolP("all-runs", "a", false, "Check all runs of the task.")
	statusCmd.Flags().IntP("run", "r", -1, "Specifies which run to consider.")

	artifactsCmd := &cobra.Command{
		Use:   "artifacts <taskId>",
		Short: "Get the name of the artifacts of a task.",
		RunE:  executeHelperE(runArtifacts),
	}
	artifactsCmd.Flags().IntP("run", "r", -1, "Specifies which run to consider.")

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
		// group
		&cobra.Command{
			Use:   "group <taskId>",
			Short: "Get the groupID of a task.",
			RunE:  executeHelperE(runGroup),
		},
		// artifacts
		artifactsCmd,
	)

	// Commands that take actions
	Command.AddCommand(
		// cancel
		&cobra.Command{
			Use:   "cancel <taskId>",
			Short: "Get the groupID of a task.",
			RunE:  executeHelperE(runCancel),
		},
		// cancel
		&cobra.Command{
			Use:   "rerun <taskId>",
			Short: "Reruns a task.",
			RunE:  executeHelperE(runRerun),
		},
		// cancel
		&cobra.Command{
			Use:   "complete <taskId>",
			Short: "Completes the execution of a task.",
			RunE:  executeHelperE(runComplete),
		},
	)

	// Add the task subtree to the root.
	root.Command.AddCommand(Command)
}

func executeHelperE(f Executor) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var creds *tcclient.Credentials
		if config.Credentials != nil {
			creds = config.Credentials.ToClientCredentials()
		}

		if len(args) < 1 {
			return fmt.Errorf("%s expects argument <taskId>", cmd.Name())
		}
		return f(creds, args, cmd.OutOrStdout(), cmd.Flags())
	}
}
