package cmd

import (
	"errors"
	"testing"

	"github.com/sneakybeaky/ebs-volumes/shared"
)

func TestErrorReturnedWhenNoInstanceFound(t *testing.T) {

	saved := getInstance
	defer func() {
		getInstance = saved
	}()

	getInstance = func() (*shared.EC2Instance, error) {
		return nil, errors.New("No instance available")
	}

	err := infoCmd.Execute()

	if err == nil {
		t.Error("No error returned")
	}
}
