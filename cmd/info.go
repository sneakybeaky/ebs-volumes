package cmd

import (
	"fmt"

	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about volumes and setup",
	Long:  `Shows the volumes assigned, thier status and detach setup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printInfo()
	},
}

func printInfo() error {

	instance, err := shared.GetInstance()

	if err != nil {
		return fmt.Errorf("Unable to get EC2 instance : %v", err)
	}

	return instance.ShowVolumesInfo()

}
