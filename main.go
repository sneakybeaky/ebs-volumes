package main

import (
	"github.com/sneakybeaky/ebs-volumes/cmd"
	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/sneakybeaky/ebs-volumes/shared/log"
)

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

	cmd.Execute()

}
