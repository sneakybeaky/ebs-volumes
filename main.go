package main

import "flag"
import (
	"fmt"
	"errors"
	"os"
)

type action struct {
	action string
	set bool
}

func (action *action) String() string {
	return action.action
}

func (action *action) Set(value string) error {

	if action.action != "" {
		return errors.New("The action has already been set")
	}

	switch value {
	case
		"attach",
		"detach",
		"info":
		action.set = true
		action.action = value
		return nil
	}
	return fmt.Errorf("Unrecognised action '%s'",action)

}

var actionFlag action

func init() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Actions on ebs volumes assigned to this instance\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Var(&actionFlag, "action", "One of attach, detach or info")
}

func main() {

	flag.Parse()

	if actionFlag.set == false {
		flag.Usage()
	}
}
