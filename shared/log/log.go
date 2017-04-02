package log

import (
	"io/ioutil"
	"log"
	"os"
)

// Debug, Info and Error are used for logging at the appropriate level

var (
	Debug = log.New(ioutil.Discard, "ebs-volumes: ", 0)
	Info  = log.New(os.Stdout, "ebs-volumes: ", 0)
	Error = log.New(os.Stdout, "ebs-volumes: ", 0)
)

// SetVerbose turns on logging categories for verbose logging
func SetVerbose() {
	Debug.SetOutput(os.Stdout)
}
