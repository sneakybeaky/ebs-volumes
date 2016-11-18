package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
	Debug = log.New(os.Stdout, "DEBUG ", log.LstdFlags)
	Info  = log.New(os.Stdout, "INFO ", log.LstdFlags)
	Error = log.New(os.Stdout, "ERROR ", log.LstdFlags)
)
