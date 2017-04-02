package cmd

import (
	"errors"
	"testing"

	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/spf13/cobra"
)

var commandtests = []struct {
	command     *cobra.Command
	description string
}{
	{infoCmd, "infoCmd"},
	{detachCmd, "detachCmd"},
	{attachCmd, "attachCmd"},
}

func TestCommandErrorsWhenNoInstanceFound(t *testing.T) {

	for _, tt := range commandtests {

		saved := getInstance
		defer func() {
			getInstance = saved
		}()

		getInstance = func() (*shared.EC2Instance, error) {
			return nil, errors.New("No instance available")
		}

		err := tt.command.Execute()

		if err == nil {
			t.Errorf("No error returned for %s when no EC2 instance found", tt.description)
		}

	}

}
