package cmd

import (
	"fmt"
	"os"

	"github.com/sneakybeaky/ebs-volumes/shared/log"
	"github.com/spf13/cobra"
)

var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "ebs-volumes",
	Short: "Attaches AWS EBS volumes to an EC2 instance",
	Long: `ebs-volumes uses tags set against an EC2 instance to attach and detach EBS volumes
To designate volumes allocated to an EC2 instance set tags with the following syntax

	volume_<device_name>=<volume_id>

For example,

	volume_/dev/sdg=vol-049df61146c4d7901

To signal that volumes should be detached set the following tag

	detach_volumes=true`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetVerbose()
		}
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(infoCmd)
	RootCmd.AddCommand(attachCmd)
	RootCmd.AddCommand(detachCmd)

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
