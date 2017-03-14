package task

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/taskcluster/taskcluster-cli/root"
	"github.com/taskcluster/taskcluster-client-go/codegenerator/model"

	"github.com/spf13/cobra"
)

var (
	pingURLs map[string]string
)

func init() {
	err := validateCache()
	if err != nil {
		log.Fatal(err)
	}
	validArgs := make([]string, len(pingURLs))
	i := 0
	for k := range pingURLs {
		validArgs[i] = k
		i++
	}
	use := "status"
	for _, validArg := range validArgs {
		use = use + " [" + validArg + "]"
	}
	statusCmd := &cobra.Command{
		Short: "taskcluster-cli status will query the current running status of taskcluster services",
		Long: `When called without arguments, taskcluster-cli status will return the current running
status of all production taskcluster services.

By specifying one or more optional services as arguments, you can limit the
services included in the status report.`,
		PreRunE:            preRun,
		Use:                use,
		ValidArgs:          validArgs,
		RunE:               status,
		DisableFlagParsing: true,
	}

	// Add the task subtree to the root.
	root.Command.AddCommand(statusCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return validateArgs(cmd, args)
}

func validateCache() error {
	return fetchManifest("https://references.taskcluster.net/manifest.json")
}

func fetchManifest(manifestURL string) error {
	var allAPIs map[string]string
	err := objectFromJsonURL(manifestURL, &allAPIs)
	if err != nil {
		return err
	}
	pingURLs = map[string]string{}
	for _, apiURL := range allAPIs {
		reference := new(model.API)
		err := objectFromJsonURL(apiURL, reference)
		if err != nil {
			return err
		}

		// loop through entries to find a /ping endpoint
		for _, entry := range reference.Entries {
			if entry.Name == "ping" {
				// determine hostname
				u, err := url.Parse(reference.BaseURL)
				if err != nil {
					return err
				}
				hostname := u.Hostname()
				//			log.Printf("URL: %v", reference.BaseURL)
				service := strings.SplitN(hostname, ".", 2)[0]
				pingURLs[service] = reference.BaseURL + entry.Route
				log.Printf("URL: %v", pingURLs[service])
				//loop through entries to get the status

				break
			}
		}
	}
	return nil
}

func objectFromJsonURL(urlReturningJSON string, object interface{}) (err error) {
	//log.Printf("Reading from %v", urlReturningJSON)
	resp, err := http.Get(urlReturningJSON)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("Bad (!= 200) status code %v from (*URL) Hostnamerl %v", resp.StatusCode, urlReturningJSON)
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&object)
	if err != nil {
		return err
	}
	return nil
}

func validateArgs(cmd *cobra.Command, args []string) error {
outer:
	for _, arg := range args {
		for _, validArg := range cmd.ValidArgs {
			if arg == validArg {
				continue outer
			}
		}
		return fmt.Errorf("invalid argument(s) passed")
	}
	return nil
}

func status(cmd *cobra.Command, args []string) error {
	//	if (len)args==0{
	//.....
	//	}
	for _, service := range args {
		fmt.Printf("%v status: %v\n", service, "running")
	}
	return nil
}
