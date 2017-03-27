package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "N/A"
	BuildTime = "N/A"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Long:  `Various version details`,
	RunE: func(cmd *cobra.Command, args []string) error {
		printVersion()
		return nil
	},
}

func printVersion() {
	fmt.Fprintf(os.Stdout, "Version %s built on %s\n", Version, BuildTime)
}
