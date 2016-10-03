package s3

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/docopt/docopt-go"
	"github.com/stretchr/testify/assert"
	"github.com/taskcluster/taskcluster-cli/client"
	"github.com/taskcluster/taskcluster-cli/config"
	"github.com/taskcluster/taskcluster-cli/extpoints"
	"github.com/taskcluster/taskcluster-cli/version"
)

func TestPutFolder1Folder2(t *testing.T) {
	// Initial setup
	var s3Test S3

	// Create a file to upload
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error occured: ", err)
	}

	// Create filepath
	filename := filepath.Join(pwd, "testFile.txt")
	if err := os.MkdirAll(filepath.Dir(filename), 0775); err != nil {
		fmt.Println("Unable to create directory: ", err)
	}

	// Setup the local file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write to file
	data := []byte("Hello World")
	file.Write(data)

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "put", "testFile.txt", "test-bucket-for-any-garbage", "folder1/folder2/"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	// Load configuration
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration file, error: ", err)
	}

	// context given to Execute
	var context extpoints.Context
	context.Arguments = arguments
	context.Config = nil
	context.Credentials = &client.Credentials{
		ClientID:    config["config"]["clientId"].(string),
		AccessToken: config["config"]["accessToken"].(string),
		Certificate: config["config"]["certificate"].(string),
	}

	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}

func TestGetFolder1Folder2(t *testing.T) {
	// Initial Setup
	var s3Test S3

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "get", "testFile.txt", "test-bucket-for-any-garbage", "folder1/folder2/"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	// Load configuration
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration file, error: ", err)
	}

	// context given to Execute
	var context extpoints.Context
	context.Arguments = arguments
	context.Config = nil
	context.Credentials = &client.Credentials{
		ClientID:    config["config"]["clientId"].(string),
		AccessToken: config["config"]["accessToken"].(string),
		Certificate: config["config"]["certificate"].(string),
	}

	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}

func TestGetFolder1(t *testing.T) {
	// Initial Setup
	var s3Test S3

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "get", "testFile.txt", "test-bucket-for-any-garbage", "folder1/"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	// Load configuration
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration file, error: ", err)
	}

	// context given to Execute
	var context extpoints.Context
	context.Arguments = arguments
	context.Config = nil
	context.Credentials = &client.Credentials{
		ClientID:    config["config"]["clientId"].(string),
		AccessToken: config["config"]["accessToken"].(string),
		Certificate: config["config"]["certificate"].(string),
	}

	result := s3Test.Execute(context)

	assert.Equal(t, false, result, "Expected an error.")
}
