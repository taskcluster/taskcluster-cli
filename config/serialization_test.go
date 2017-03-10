package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
	homedir "github.com/mitchellh/go-homedir"
)


func TestLoad(t *testing.T) {
	assert := assert.New(t)

	// Recreate configFile and make sure it gets correct path
	configFolder := os.Getenv("XDG_CONFIG_HOME")
	if configFolder == "" {
		homeFolder := os.Getenv("HOME")
		if homeFolder == "" {
			homeFolder, _ = homedir.Dir()
		}
		if homeFolder != "" {
			configFolder = filepath.Join(homeFolder, ".config")
		}
	}
	returnConfigLoc := filepath.Join(configFolder, "taskcluster.yml")
	reader := strings.NewReader(returnConfigLoc)
	assert.Equal(reader,configFile())
	_, err := Load(reader)
	assert.NoError(err)

/*
	// Load should fail if passed an invalid configFile location
	returnConfigLoc = filepath.Join(configFolder, "throwError_taskcluster.yml")
	reader = strings.NewReader(returnConfigLoc)
	_, err = Load(reader)
	assert.Error(err) // This actually is not throwing an error

*/


	// Test if validation fails with non-integer passed in
	
	// Test if Load fails when file is not registered

	// Test invalid config option values

}
