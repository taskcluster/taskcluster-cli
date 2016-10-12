package download

import (
	"log"
	"time"

	"github.com/alexandrasp/taskcluster-cli/extpoints"
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
			url_artifact, err := userQueue.GetArtifact_SignedURL(taskId, runId, artifact, time.Second*300)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {

			}
		}
		if runId == "" {
			//get latest artifact without rundId parameter
			url_artifact, err := userQueue.GetLatestArtifact_SignedURL(taskId, artifact, time.Second*300)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {

			}
		}

	}
	return false
}
