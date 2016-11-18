package expandscope

import (
	"fmt"
	"strings"

	"github.com/taskcluster/taskcluster-cli/extpoints"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func init() {
	extpoints.Register("expand-scope", expandscope{})
}

type expandscope struct{}

func (expandscope) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (expandscope) Summary() string {
	return "Expand the given scope set."
}

func usage() string {
	return `Usage:
  taskcluster expand-scope <scope>...

This command returns an expanded copy of the given scope set, with scopes
implied by any roles included. The given scope set is specified as a space
separated list of scopes.
`
}

func (expandscope) Usage() string {
	return usage()
}

func (expandscope) Execute(context extpoints.Context) bool {
	argv := context.Arguments
	inputScopes := argv["<scope>"].([]string)

	if argv["expand-scope"].(bool) {
		fmt.Printf("%s\n", expandScope(inputScopes))
		return true
	}
	return true
}

func expandScope(inputScopes []string) string {

	a := auth.New(&tcclient.Credentials{})
	a.Authenticate = false

	params := &auth.SetOfScopes{
		Scopes: inputScopes,
	}

	resp, err := a.ExpandScopes(params)
	if err != nil {
		fmt.Printf("Error expanding scopes: %s\n", err)
	}

	expandedScope := strings.Join(resp.Scopes, "\n")
	return expandedScope
}
