package task

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	assert "github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tcclient "github.com/taskcluster/taskcluster-client-go"
)

const fakeTaskID = "ANnmjMocTymeTID0tlNJAw"

type FakeServerSuite struct {
	suite.Suite
	testServer *httptest.Server
}

func (suite *FakeServerSuite) SetupSuite() {
	// set up a fake server that knows how to answer the `task()` method
	handler := http.NewServeMux()
	handler.HandleFunc("/v1/task/"+fakeTaskID,
		func(w http.ResponseWriter, _ *http.Request) {
			io.WriteString(w, `{"metadata": {"name": "my-test"}}`)
		})
	suite.testServer = httptest.NewServer(handler)

	// set the base URL the subcommands use to point to the fake server
	queueBaseURL = suite.testServer.URL + "/v1"
}

func (suite *FakeServerSuite) TearDownSuite() {
	suite.testServer.Close()
	queueBaseURL = ""
}

func (suite *FakeServerSuite) TestNameCommand() {
	// set up to run a command and capture output
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOutput(buf)

	// run the command
	args := []string{fakeTaskID}
	runName(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal(string(buf.Bytes()), "my-test\n")
}

func TestFakeServerSuite(t *testing.T) {
	suite.Run(t, new(FakeServerSuite))
}

// Test the `task log` subcommand against a real task, since it does its own
// HTTP handling
func TestLogCommand(t *testing.T) {
	assert := assert.New(t)

	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOutput(buf)

	args := []string{"TtAsnXdCS1-tAQxvMO4rHQ"}
	runLog(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	// This is the output of a static log
	s := "[taskcluster 2017-03-03 21:18:34.946Z] Task ID: TtAsnXdCS1-tAQxvMO4rHQ\n" +
		"[taskcluster 2017-03-03 21:18:34.946Z] Worker ID: i-035dd1bd8da876f13\n" +
		"[taskcluster 2017-03-03 21:18:34.946Z] Worker Group: us-west-1b\n" +
		"[taskcluster 2017-03-03 21:18:34.946Z] Worker Node Type: m3.large\n" +
		"[taskcluster 2017-03-03 21:18:34.946Z] Worker Type: tutorial\n" +
		"[taskcluster 2017-03-03 21:18:34.946Z] Public IP: 54.153.13.193\n" +
		"\n" +
		"[taskcluster 2017-03-03 21:18:48.518Z] === Task Starting ===\n" +
		"hello World\n" +
		"[taskcluster 2017-03-03 21:18:48.945Z] === Task Finished ===\n" +
		"[taskcluster 2017-03-03 21:18:48.946Z] Successful task run with exit code: 0 completed in 14.001 seconds\n"

	assert.Equal(string(buf.Bytes()), s, "Log's are not equal.")
}
