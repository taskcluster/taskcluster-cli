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

// Initial setup
var s3Test S3
var context extpoints.Context

func createFile() string {
	// Get current path
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error occured: ", err)
	}

	// Create filepath
	filename := filepath.Join(pwd, "testFile.txt")

	// Setup the local file
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write to file
	data := []byte("Hello World")
	file.Write(data)
	return filename
}

func loadConfig() {
	// Load configuration
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration file, error: ", err)
	}

	// Set context
	context = extpoints.Context{}
	context.Config = nil
	context.Credentials = &client.Credentials{
		ClientID:    config["config"]["clientId"].(string),
		AccessToken: config["config"]["accessToken"].(string),
		Certificate: config["config"]["certificate"].(string),
	}
}

func TestPutFolder1Folder2(t *testing.T) {
	filename := createFile()
	loadConfig()

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "put", filename, "test-bucket-for-any-garbage", "folder1/folder2/"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	context.Arguments = arguments
	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}

func TestGetWithTarget(t *testing.T) {
	// Get current path
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error occured: ", err)
	}
	target := filepath.Join(pwd, "testFile.txt")

	loadConfig()

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "get", "test-bucket-for-any-garbage", "folder1/folder2/testFile.txt", target}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	context.Arguments = arguments
	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}

func TestGetWithOutTarget(t *testing.T) {
	loadConfig()

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "get", "test-bucket-for-any-garbage", "folder1/folder2/testFile.txt"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	context.Arguments = arguments
	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}

func TestGetFolder1(t *testing.T) {
	loadConfig()

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "get", "test-bucket-for-any-garbage", "folder1/testFile.txt"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	context.Arguments = arguments
	result := s3Test.Execute(context)

	assert.Equal(t, false, result, "Expected an error.")
}

func TestPutToFile(t *testing.T) {
	filename := createFile()
	loadConfig()

	// Parse arguments
	os.Args = []string{"taskcluster", "s3", "put", filename, "test-bucket-for-any-garbage", "folder1/folder2/testFile.txt"}
	arguments, err := docopt.Parse(usage(), nil, true, version.VersionNumber, true)
	if err != nil {
		fmt.Println("Failed to parse arguments")
	}

	context.Arguments = arguments
	result := s3Test.Execute(context)

	assert.Equal(t, true, result, "An error occured during execution.")
}
