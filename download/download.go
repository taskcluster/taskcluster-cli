package download

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alexandrasp/taskcluster-cli/extpoints"
	"github.com/taskcluster/httpbackoff"
	"github.com/taskcluster/taskcluster-client-go"
	"github.com/taskcluster/taskcluster-client-go/queue"
)

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
			<taskID> [<runID>] <artifact>
			`
}

func (download) Usage() string {
	return usageDownload()
}

func (download) Execute(context extpoints.Context) bool {
	command := context.Arguments["download"].(string)
	taskID := context.Arguments["<taskID>"].(string)
	runID := context.Arguments["<runID>"].(string)
	artifact := context.Arguments["<artifact>"].(string)
	provider := extpoints.CommandProviders()[command]
	if provider == nil {
		log.Panicf("Unknown command %s", command)
	} else {
		permaCred := &tcclient.Credentials{
			ClientID:    os.Getenv("TASKCLUSTER_CLIENT_ID"),
			AccessToken: os.Getenv("TASKCLUSTER_ACCESS_TOKEN"),
		}
		urlArtifact, err := getSignedUrlArtifact(taskID, runID, artifact, permaCred)
		if err != nil {
			log.Panicf("Exception thrown signing URL \n%s", err)
		} else {
			urlArtifact := EnforceHttpsUrl(urlArtifact.String())
			response, attempts, err := getArtifact(urlArtifact)
			if err != nil {
				log.Panicf("Exception thrown downloading an artifact \n%s", err)
			} else {
				fmt.Printf("Number of attempts: %d\n", attempts)
				length, out, _ := checkContentLength(response)
				verifyContentTypeUncompress(response)
				log.Printf("ContentLength %d with %s\n", length, out)
				if StreamingArtifactFile(response, taskID) {
					fmt.Print("Save with sucess")
				} else {
					fmt.Print("Error to save a file")
				}
			}
		}
	}
	return true
}

//fuction to download a artifact with automatic retries
func getArtifact(url string) (*http.Response, int, error) {
	res, attempts, err := httpbackoff.Retry(func() (*http.Response, error, error) {
		resp, err := http.Get(url)
		// assume all errors are temporary
		//following redirect if there is a new url, link, redirect
		return resp, err, nil
	})
	return res, attempts, err
}

//function to get a signed URL to an artifact
func getSignedUrlArtifact(taskID string, runID string, artifact string, permaCred *tcclient.Credentials) (*url.URL, error) {
	userQueue := queue.New(permaCred)
	if runID != "" {
		//get an artifact with runID parameter
		urlArtifact, err := userQueue.GetArtifact_SignedURL(taskID, runID, artifact, time.Second*10)
		return urlArtifact, err
	} else {
		//get latest artifact without rundId parameter
		urlArtifact, err := userQueue.GetLatestArtifact_SignedURL(taskID, artifact, time.Second*10)
		return urlArtifact, err
	}
}

//func to check an artifact content
func checkContentLength(res *http.Response) (int64, string, error) {
	if res.ContentLength > 0 {
		return res.ContentLength, "Good", nil
	}
	if res.ContentLength == 0 {
		//Means exactly none
		if res.Body != nil {
			return res.ContentLength, "Artifact content is empty", nil
		}
		return res.ContentLength, "None", nil
	}
	if res.ContentLength < 0 {
		//Means Unknown
		return res.ContentLength, "Chunked", nil
	}
	return 0, "", nil
}

//fuction to stream response struct to an output file
func StreamingArtifactFile(res *http.Response, taskID string) bool {
	fmt.Print("BODY:::", res.Body)
	body, err := ioutil.ReadAll(res.Body)
	err = ioutil.WriteFile(taskID+".txt", body, 0644)
	if err != nil {
		return false
	} else {
		return true
	}
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

//function to uncompress an artifact if it is bzip
func verifyContentTypeUncompress(resp *http.Response) {
	if strings.Index(resp.Header.Get("Content-Type"), "x-bzip2") != -1 {
		fmt.Print(resp.Header.Get("Content-Encoding"))
	}
	if strings.Index(resp.Header.Get("Content-Type"), "zip") != -1 {
		fmt.Print(resp.Header.Get("Content-Encoding"))
	}
}
