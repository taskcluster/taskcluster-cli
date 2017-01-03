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
	return "Shows whether a given scope or set of scope satifies another."
}

func (scopecheck) Usage() string {
	return `Usage:
  taskcluster scope-check <scope1> satisfies <scope2>
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
	satisfies := argv["satisfies"]
	_ = satisfies
	scope2 := argv["<scope2>"].(string)
	response := checkscopes(scope1, scope2)
	fmt.Printf("%s\n", response)

	return true
}

func checkscopes(scope1 string, scope2 string) string {
	var secondScope interface{} = scope2
	if strings.HasPrefix("scope2", "assume:") {
		//cast scope2 to string array before calling api method expandingScope
		expandedScope, errs := expandScope(secondScope.([]string))
		if errs != nil {
			resp := "Error while trying to expand scopes"
			return resp
		}
		//Need it back to string for comparison with scope1
		var expdscope interface{} = expandedScope
		scope2 = expdscope.(string)

	} else {

		scope2 = secondScope.(string)
	}

	if scope1 == scope2 {
		return "YES"
	} else {
		resp := "NO missing -"
		return resp
	}

}
