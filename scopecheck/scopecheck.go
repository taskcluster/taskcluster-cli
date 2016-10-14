package scopecheck

import (
	"fmt"

	"github.com/taskcluster/taskcluster-cli/extpoints"
	tcclient "github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/auth"
	scopeutility "github.com/taskcluster/taskcluster-lib-scopes"
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
  taskcluster scope-check <scope1> satifies <scope2>
`
}

func expandScope(Auth *auth.Auth, scope2 string) (string, error) {

	scopeParams := &auth.SetOfScopes{
		Scopes: scope2,
	}

	response, errors := Auth.ExpandScopes(scopeParams)
	if errors != nil {
		fmt.Printf("Error expanding scopes: %s\n", err)
		return false
	}

	_, expandedScopes := response.Scopes
	return expandedScopes.(string), nil
}

func (scopecheck) Execute(context extpoints.Context) bool {
	argv := context.Arguments

	command := argv["SCOPECHECK"].(string)
	provider := extpoints.CommandProviders()[command]
	if provider == nil {
		panic(fmt.Sprintf("Unknown command: %s", command))
		return false
	}

	satisfies := argv("<satisfies>")
	scope1 := argv("<scope1>").([]string)
	scope2 := argv("<scope2>").([]string)

	if argv["scope1"].(bool) {

		response := checkscopes(argv, scope1, scope2)
		fmt.Printf("%s\n", response)
	}
	return true
}

func checkscopes(scope1 string, scope2 string) string {

	auth := auth.New(&tcclient.Credentials{})
	if scope1 == scope2 {
		return "YES"
	} else {

		scope, errs := expandScope(auth, scope2)
		if errs != nil {
			resp := "Error while trying to expand scopes"
			return resp
		}

		if scopeutility.scopeMatch(scope1, scope) {
			resp := "YES"
			return resp
		} else {
			resp := "NO missing -"
			resp += scope2
			return resp
		}

	}

}
