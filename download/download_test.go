package download

import (
	"testing"

	"github.com/alexandrasp/taskcluster-cli/extpoints"
)

func TestSummary(t *testing.T) {
	dl := download{}
	expectedStr := "Download an artifact"
	result := dl.Summary()
	if result != expectedStr {
		t.Fatalf("Expected %s got %s", expectedStr, result)
	}
}

func TestUsage(t *testing.T) {
	dl := download{}
	result := dl.Usage()
	t.Log(result)
}

func TestInit(t *testing.T) {
	dl := download{}
	result := dl.ConfigOptions()
	t.Log(result)
}

func TestExecute(t *testing.T) {
	dl := download{}
	myContext := extpoints.Context{}
	myContext.Arguments = make(map[string]interface{})
	myContext.Arguments["download"] = "download"
	myContext.Arguments["<taskID>"] = "LvRW86t_QpyahfuE60GW_A"
	myContext.Arguments["<runID>"] = ""
	myContext.Arguments["<artifact>"] = "public/build/target.reftest.tests.zip"
	response := dl.Execute(myContext)
	t.Log(response)
}
