package cmd

import (
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

func showVolumesInfo(instance *shared.EC2Instance) error {
	return instance.ShowVolumesInfo()
}
