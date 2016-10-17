package download

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
			ClientID:    os.Getenv("TASKCLUSTER_CLIENT_ID"),
			AccessToken: os.Getenv("TASKCLUSTER_ACCESS_TOKEN"),
		}

		userQueue := queue.New(permaCred)

		if runId != "" {
			//get a artifact with runId parameter
			url_artifact, err := userQueue.GetArtifact_SignedURL(taskId, runId, artifact, time.Second*10)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {

				url_artifact := EnforceHttpsUrl(url_artifact.String())
				fmt.Printf("NEW URL %s\n", url_artifact)
				response, attempts, err := getAnArtifact(url_artifact)

				if err != nil {
					log.Panicf("Exception thrown download an artifact \n%s", err)
				} else {
					fmt.Printf("Number of attempts: %d\n", attempts)
					_, length, out := checkContentLength(response)
					log.Printf("ContentLength %d with %s", length, out)
				}
			}
		}
		if runId == "" {
			//get latest artifact without rundId parameter
			url_artifact, err := userQueue.GetLatestArtifact_SignedURL(taskId, artifact, time.Second*10)
			if err != nil {
				log.Panicf("Exception thrown signing URL \n%s", err)
			} else {

				url_artifact := EnforceHttpsUrl(url_artifact.String())
				//fmt.Printf("NEW URL %s\n", url_artifact)
				response, attempts, err := getAnArtifact(url_artifact)

				if err != nil {
					log.Panicf("Exception thrown download an artifact \n%s", err)
				} else {
					fmt.Printf("Number of attempts: %d\n", attempts)
					fmt.Printf("Content:::: \n %s \n", response.Header.Get("Content-Type"))
					_, length, out := checkContentLength(response)
					log.Printf("ContentLength %d with %s\n", length, out)
				}
			}
		}

	}
	return true
}

func getAnArtifact(url string) (*http.Response, int, error) {
	res, attempts, err := httpbackoff.Retry(func() (*http.Response, error, error) {
		resp, err := http.Get(url)
		// assume all errors are temporary

		//following redirect if there is a new url, link, redirect
		return resp, err, nil
	})
	return res, attempts, err
}

func checkContentLength(res *http.Response) (error, int64, string) {

	if res.ContentLength > 0 {
		return nil, res.ContentLength, "Good"
	}
	if res.ContentLength == 0 {
		//Means exactly none
		if res.Body != nil {
			return nil, res.ContentLength, "None With Some Body Content"
		}
		return nil, res.ContentLength, "None"
	}
	if res.ContentLength < 0 {
		//Means Unknown
		return nil, res.ContentLength, "Chunked"
	}
	return nil, 0, ""
}

func StreamingArtifactFile(res *http.Response) bool {

	return false

}

//function to enforce https if an artifact came without https assigned
func EnforceHttpsUrl(url string) string {

	indexHttps := strings.Index(url, "https://")
	if indexHttps == 0 {
		return url
	}
	if indexHttps == -1 {
		indexHttp := strings.Index(url, "http://")
		if indexHttp == 0 {
			url = strings.Replace(url, "p", "ps", 3)
		} else {
			s := []string{"https://", url}
			url = strings.Join(s, "")
		}
	}
	return url
}
