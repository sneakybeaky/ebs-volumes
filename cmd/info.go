package cmd

import (
	"fmt"

	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about volumes and setup",
	Long:  `Shows the volumes assigned, their status and detach setup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return apply(showVolumesInfo)
	},
}

func apply(action func(*shared.EC2Instance) error) error {

	instance, err := getInstance()

	if err != nil {
		return fmt.Errorf("unable to get EC2 instance : %v", err)
	}

	return action(instance)

}

var getInstance = func() (*shared.EC2Instance, error) {
	return shared.GetInstance()
}

func showVolumesInfo(instance *shared.EC2Instance) error {
	return instance.ShowVolumesInfo()
}
