package main

import (
	"fmt"
	"os"

	"github.com/taskcluster/taskcluster-cli/config"
	"github.com/taskcluster/taskcluster-cli/cmds/root"
)

func main() {
	// set up the whole config thing
	var file *os.File
	var err error
	if file, err = config.ConfigFile(os.O_RDONLY); err != nil{
		fmt.Fprintf(os.Stderr, "failed to open configuration file, error: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()
	config.Setup(file)

	// gentlemen, START YOUR ENGINES
	if err := root.Command.Execute(); err != nil {
		os.Exit(1)
	} else {

		os.Exit(0)
	}
}
