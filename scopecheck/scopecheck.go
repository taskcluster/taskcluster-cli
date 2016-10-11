package scopecheck

import (
	"fmt"

	"github.com/taskcluster/taskcluster-cli/extpoints"
	tcclient "github.com/taskcluster/taskcluster-client-go"
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

func usage() string {
	return `Usage:
  taskcluster scope-check <scope1> satifies <scope2>
`
}

func (scopecheck) Usage() string {
	return usage()
}

type Auth tcclient.ConnectionData

func New(credentials *tcclient.Credentials) *Auth {
	myAuth := Auth(tcclient.ConnectionData{
		Credentials:  credentials,
		BaseURL:      "https://auth.taskcluster.net/v1",
		Authenticate: true,
	})
	return &myAuth
}

//may have duplicate of this method because this method will be written at some point in issue#31
func (myAuth *Auth) expandScope(payload *SetOfScopes) (*SetOfScopes, error) {
	connectionDetails := tcclient.ConnectionData(*myAuth)
	responseObject, _, err := (&connectionDetails).APICall(payload, "GET", "/scopes/expand", new(SetOfScopes), nil)
	return responseObject.(*SetOfScopes), err
}

func (scopecheck) Execute(context extpoints.Context) bool {
	argv := context.Arguments

	command := argv["scope-check"].(string)
	provider := extpoints.CommandProviders()[command]
	if provider == nil {
		panic(fmt.Sprintf("Unknown command: %s", command))
		return false
	}

	satisfies := argv("<satisfies>")
	rscope := argv("<rscope>")
	lscope := argv("<lscope>")

	if argv["rscope"].(bool) {

		response := checkscopes(rscope, lscope)
		fmt.Printf("%s\n", response)
	}
	return true
}

func checkscopes(rightScope *SetOfScopes, leftScope *SetOfScopes) string {

	if rightScope == leftScope {

		return "YES"
	} else {

		scope, errs := expandScope(leftScope)
		if errs != nil {
			resp := "Error while trying to expand scopes"
			return resp
		}

		if scopeutility.scopeMatch(rightScope, scope) {
			resp := "YES"
			return resp
		} else {
			resp := "NO missing -"
			resp += leftScope
			return resp
		}

	}

}
