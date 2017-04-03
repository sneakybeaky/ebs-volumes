package cmd

import (
	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach volumes",
	Long:  `Attaches volumes designated via tags`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return apply(attachVolumes)
	},
}

func attachVolumes(instance *shared.EC2Instance) error {
	return instance.AttachVolumes()
}
