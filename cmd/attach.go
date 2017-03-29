package cmd

import (
	"fmt"

	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach volumes",
	Long:  `Attaches volumes designated via tags`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return attach()
	},
}

func attach() error {

	instance, err := shared.GetInstance()

	if err != nil {
		return fmt.Errorf("Unable to get EC2 instance : %v", err)
	}

	return instance.AttachVolumes()
}
