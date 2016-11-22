package shared_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared"
	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {

			if *input.VolumeIds[0] == expectedVolumeID {
				return nil
			}

			return fmt.Errorf("Unexpected volume id %s", *input.VolumeIds[0])

		},
	}

	underTest := shared.NewAllocatedVolume("vol-54321", "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err != nil {
		t.Errorf("Detaching the volume shouldn't have failed, but I got %v", err)
	}
}
