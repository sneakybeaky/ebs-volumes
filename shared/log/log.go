package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Debug = log.New(ioutil.Discard, "", 0)
	Info  = log.New(os.Stdout, "", 0)
	Error = log.New(os.Stdout, "", 0)
)

func SetVerbose() {
	Debug.SetOutput(os.Stdout)
}
