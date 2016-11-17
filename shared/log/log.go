package log
import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Debug = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
	Info = log.New(os.Stdout, "INFO ", log.LstdFlags)
	Error = log.New(os.Stdout, "ERROR ", log.LstdFlags)
)