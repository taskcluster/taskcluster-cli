package scopecheck

import (
	"fmt"
	"strings"

	"github.com/taskcluster/taskcluster-cli/extpoints"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
)

func init() {
	extpoints.Register("scope-check", scopecheck{})
}

type scopecheck struct{}

func (scopecheck) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (scopecheck) Summary() string {
	return "Shows whether a given scope satifies another."
}

func (scopecheck) Usage() string {
	return `Usage:
  taskcluster scope-check <scope1> <scope2>
`
}

func expandScope(scope2 []string) (*auth.SetOfScopes, error) {

	a := auth.New(&tcclient.Credentials{})
	a.Authenticate = false

	scopes := &auth.SetOfScopes{
		Scopes: scope2,
	}

	resp, err := a.ExpandScopes(scopes)
	if err != nil {
		fmt.Printf("Error expanding scopes: %s\n", err)
		return nil, err
	}

	return resp, err
}

func (scopecheck) Execute(context extpoints.Context) bool {
	argv := context.Arguments
	scope1 := argv["<scope1>"].(string)
	scope2 := argv["<scope2>"].(string)

	if argv["scope-check"].(bool) {
		response := checkscopes(scope1, scope2)
		fmt.Printf("%s\n", response)
	}
	return true
}

func checkscopes(scope1 string, secondScope string) string {

	var scope2 string
	var containerforScope2 interface{} = secondScope
	if strings.HasPrefix("secondScope", "assume:") {
		expandedScope, errs := expandScope(containerforScope2.([]string))
		if errs != nil {
			resp := "Error while trying to expand scopes"
			return resp
		}
		var containerforScope2 interface{} = expandedScope
		scope2 = containerforScope2.(string)

	} else {

		scope2 = containerforScope2.(string)
	}

	if scope1 == scope2 {
		return "YES"
	} else {
		resp := "NO missing -"
		return resp
	}

}
