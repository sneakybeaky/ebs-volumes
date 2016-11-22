package shared

import (
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
	"github.com/aws/aws-sdk-go/private/waiter"
	"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
	}

	revert := waitForVolumeIDSucceeds(expectedVolumeID)
	defer revert()

	underTest := NewAllocatedVolume("vol-54321","/dev/sdg","i-11223344",mockEC2Service)

	err := underTest.Detach()

	if err != nil {
		t.Errorf("Detaching the volume shouldn't have failed, but I got %v", err)
	}
}

func waitForVolumeIDSucceeds(volumeID string) func() {

	saved := invokeWait

	invokeWait = func(waiter waiter.Waiter) error{
		input := waiter.Input.(*ec2.DescribeVolumesInput)

		if *input.VolumeIds[0] != volumeID {
			return fmt.Errorf("I was expecting %s but got %s",volumeID,*input.VolumeIds[0])
		} else {
			return nil
		}

	}

	return func() {
		invokeWait = saved
	}

}
