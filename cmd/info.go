package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create AWS session : %v", err)
	}

	metadata := shared.NewEC2InstanceMetadata(sess)

	region, err := metadata.Region()

	if err != nil {
		return fmt.Errorf("Failed to get AWS region : %v", err)
	}

	sess.Config.Region = &region

	instance := shared.NewEC2Instance(metadata, ec2.New(sess))

	instance.ShowVolumesInfo()

	return nil
}
