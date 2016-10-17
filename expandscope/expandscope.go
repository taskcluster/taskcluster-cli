package scope

import (
	"fmt"

	"github.com/taskcluster/taskcluster-cli/extpoints"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func init() {
	extpoints.Register("scope", scope{})
}

type scope struct{}

func (scope) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (scope) Summary() string {
	return "Expand the given scope set."
}

func usage() string {
	return `Usage:
  taskcluster scope expand <scope>...

### Expand
This command returns an expanded copy of the given scope set, with scopes
implied by any roles included. The given scope set is specified as a space
separated list of scopes.
`
}

func (scope) Usage() string {
	return usage()
}

func (scope) Execute(context extpoints.Context) bool {
	argv := context.Arguments

	authCreds := tcclient.Credentials{}
	myAuth := auth.New(&authCreds)
	myAuth.Authenticate = false

	if argv["expand"].(bool) {
		return expandScope(argv, myAuth)
	}
	return true
}

func expandScope(argv map[string]interface{}, myAuth *auth.Auth) bool {
	inputScopes := argv["<scope>"].([]string)

	params := &auth.SetOfScopes{
		Scopes: inputScopes,
	}

	resp, err := myAuth.ExpandScopes(params)
	if err != nil {
		fmt.Printf("Error expanding scopes: %s\n", err)
		return false
	}

	for _, s := range resp.Scopes {
		fmt.Printf("%s\n", s)
	}
	return true
}
