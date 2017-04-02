package cmd

import (
	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach volumes",
	Long:  `Detaches volumes if enabled via tags`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return apply(detachVolumes)
	},
}

func detachVolumes(instance *shared.EC2Instance) error {
	return instance.DetachVolumes()
}
