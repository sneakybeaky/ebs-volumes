package shared_test

import (
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared"
	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc:             helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeAvailableFunc: helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID),
	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err != nil {
		t.Errorf("Detaching the volume shouldn't have failed, but I got %v", err)
	}
}

func TestAttachVolumeWhenDetached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		AttachVolumeFunc:             helper.AttachVolumeForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeAvailableFunc: helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeInUseFunc:     helper.WaitUntilVolumeInUseForVolumeIDSuccess(expectedVolumeID),
	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Attach()

	if err != nil {
		t.Errorf("Attaching the volume shouldn't have failed, but I got %v", err)
	}
}
