package main

import "flag"
import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/sneakybeaky/ebs-volumes/shared/log"
)

type action struct {
	action string
	set    bool
}

var (
	Version    = "N/A"
	BuildTime  = "N/A"
	actionFlag action
)

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
	return fmt.Errorf("Unrecognised action '%s'", action)

}

func init() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Actions on ebs volumes assigned to this instance\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Var(&actionFlag, "action", "One of attach, detach or info")
}

var showInfo = func(instance *shared.EC2Instance) {
	instance.ShowVolumesInfo()
}

var doAttach = func(instance *shared.EC2Instance) {
	instance.AttachVolumes()
}

var doDetach = func(instance *shared.EC2Instance) {
	if err := instance.DetachVolumes(); err != nil {
		log.Error.Fatalf("Unable to detach volumes : %v\n", err)
	}
}

func main() {

	versionFlag := flag.Bool("v", false, "prints current version and build")
	flag.Parse()

	if *versionFlag {
		fmt.Fprintf(os.Stderr, "Version %s built on %s\n", Version, BuildTime)
		os.Exit(0)
	}

	if actionFlag.set == false {
		flag.Usage()
		os.Exit(0)
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Error.Fatalf("failed to create session %v\n", err)
	}

	metadata := shared.NewEC2InstanceMetadata(sess)

	if region, err := metadata.Region(); err != nil {
		log.Error.Fatalf("failed to get region %v\n", err)
	} else {
		sess.Config.Region = &region
	}

	instance := shared.NewEC2Instance(metadata, ec2.New(sess))

	switch actionFlag.action {
	case "info":
		showInfo(instance)
	case "attach":
		doAttach(instance)
	case "detach":
		doDetach(instance)
	}

}
