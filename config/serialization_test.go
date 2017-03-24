package config

import (
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
    const configSample = `
                          cmd:
                          foo: bar
                         `

    options := map[string]OptionDefinition{
        "foo": OptionDefinition{},
    }
    RegisterOptions("cmd", options)
    r := strings.NewReader(configSample)
    m, err := Load(r)
    assert.NoError(t, err)
    _, ok := m["cmd"]
    assert.True(t, ok)
    _, ok = m["cmd"]["foo"]
    assert.True(t, ok)
  
}

func TestSave(t *testing.T){
	// to do 
}
