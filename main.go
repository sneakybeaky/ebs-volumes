package main

import "flag"
import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sneakybeaky/aws-volumes/shared"
	"log"
	"os"
)

type action struct {
	action string
	set    bool
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
	return fmt.Errorf("Unrecognised action '%s'", action)

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

func showInfo(instance *shared.EC2Instance) error {
	if volumes, err := instance.AllocatedVolumes(); err != nil {
		return err
	} else {
		for _, volume := range volumes {
			volume.Info(os.Stdout)
		}
	}

	return nil
}

func doAttach(instance *shared.EC2Instance) {
	if volumes, err := instance.AllocatedVolumes(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to find allocated volumes : %s", err)
	} else {

		done := make(chan error)
		defer close(done)

		for _, volume := range volumes {

			go func() {
				volume := volume
				done <- volume.Attach()
			}()

		}

		for i := 0; i < len(volumes); i++ {
			done <- err

			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to attach volume : %s", err)
			}
		}
	}

}

func main() {

	flag.Parse()

	if actionFlag.set == false {
		flag.Usage()
	}

	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create session %v\n", err)
	}

	metadata := shared.NewEC2InstanceMetadata(sess)

	if region, err := metadata.Region(); err != nil {
		log.Fatalf("failed to get region %v\n", err)
	} else {
		sess.Config.Region = &region
	}

	instance := shared.NewEC2Instance(metadata, sess)

	switch actionFlag.action {
	case "info":
		showInfo(instance)
	case "attach":
		doAttach(instance)
	}

}
