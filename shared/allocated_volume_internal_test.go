package shared

import (
	"fmt"
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
	}

	revert := waitForVolumeIDSucceeds(expectedVolumeID)
	defer revert()

	underTest := NewAllocatedVolume("vol-54321", "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err != nil {
		t.Errorf("Detaching the volume shouldn't have failed, but I got %v", err)
	}
}

func waitForVolumeIDSucceeds(volumeID string) func() {

	saved := waitUntilDetached

	waitUntilDetached = func(volume AllocatedVolume) error {

		if volume.VolumeId != volumeID {
			return fmt.Errorf("I was expecting %s but got %s", volumeID, volume.VolumeId)
		}
		return nil

	}

	return func() {
		waitUntilDetached = saved
	}

}
