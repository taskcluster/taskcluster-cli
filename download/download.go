package download

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexandrasp/taskcluster-cli/extpoints"
	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

//"github.com/taskcluster/httpbackoff"

func init() {
	extpoints.Register("download", download{})
}

type download struct{}

func (download) ConfigOptions() map[string]extpoints.ConfigOption {
	return nil
}

func (download) Summary() string {
	return "Download an artifact"
}

func usageDownload() string {
	return `Usage:
			taskcluster download [options]
			Options:
			<taskId> [<runId>] <artifact>
			`
}

func (download) Usage() string {
	return usageDownload()
}

func (download) Execute(context extpoints.Context) bool {
	command := context.Arguments["download"].(string)
	taskId := context.Arguments["<taskId>"].(string)
	runId := context.Arguments["<runId>"].(string)
	artifact := context.Arguments["<artifact>"].(string)

	provider := extpoints.CommandProviders()[command]
	if provider == nil {
		log.Panicf("Unknown command %s", command)

	} else {

		permaCred := &tcclient.Credentials{
			ClientID:    "tester",
			AccessToken: "no-secret",
		}

		userQueue := queue.New(permaCred)

		if runId != "" {
			//get a artifact with runId parameter
			url_artifact, err := userQueue.GetArtifact_SignedURL(taskId, runId, artifact, time.Second*10)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {
				getAnArtifact(url_artifact.String())
			}
		}
		if runId == "" {
			//get latest artifact without rundId parameter
			url_artifact, err := userQueue.GetLatestArtifact_SignedURL(taskId, artifact, time.Second*10)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {
				getAnArtifact(url_artifact.String())
			}
		}

	}
	return false
}

func getAnArtifact(url string) {
	res, attempts, err := httpbackoff.Retry(func() (*http.Response, error, error) {
		resp, err := http.Get(url)
		// assume all errors are temporary
		return resp, err, nil
	})

	if err != nil {

		log.Panicf("Exception thrown download an artifact \n%s", err)

	} else {

		fmt.Sscan("%d Retries", attempts)
		fmt.Printf("%+v\n", res)

	}

}
