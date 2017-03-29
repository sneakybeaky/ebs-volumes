package cmd

import (
	"fmt"

	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Detach volumes",
	Long:  `Detaches volumes if enabled via tags`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return detach()
	},
}

func detach() error {

	instance, err := shared.GetInstance()

	if err != nil {
		return fmt.Errorf("Unable to get EC2 instance : %v", err)
	}

	return instance.DetachVolumes()
}
