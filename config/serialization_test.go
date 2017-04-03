package config

import (
    "os"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestLoadAndSave(t *testing.T) {
  
    // Create tester config data and option definitions
    const configSample = `
                          cmd:
                            foo: bar
                         `
    options := map[string]OptionDefinition{
       "foo": OptionDefinition{},
    }
    
    // Register the fake config options
    RegisterOptions("cmd", options)
    
    // Call Load() with the test data
    configSampleReader := strings.NewReader(configSample)
    configMap, err := Load(configSampleReader)
    assert.NoError(t, err)
    
    // Compare Load() output and config object
    _, ok := configMap["cmd"]
    assert.True(t, ok)
    _, ok = configMap["cmd"]["foo"]
    assert.True(t, ok)

    // Test Save() with results from Load()
    configWriter,_ := os.Create(configSample)
    err = Save(configMap, configWriter)
    assert.NoError(t,err)

    // assert.Equal(configSampleReader, fileIJustMade) 

    //TODO Remove the saved file after we've tested 
    // Separate out Save and Load testing
}
