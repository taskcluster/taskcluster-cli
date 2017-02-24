package task

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
	tcclient "github.com/taskcluster/taskcluster-client-go"
)

// Executor represents the function interface of the task subcommand.
type Executor func(credentials *tcclient.Credentials, args []string, out io.Writer, flagSet *pflag.FlagSet) error

// getRunStatusString takes the state and resolved strings and crafts a printable summary string.
func getRunStatusString(state, resolved string) string {
	if resolved != "" {
		return fmt.Sprintf("%s '%s'", state, resolved)
	}

	return state
}
