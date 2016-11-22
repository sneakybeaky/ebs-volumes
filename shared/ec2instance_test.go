package shared_test

import (
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared"
	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestFindAllocatedVolumes(t *testing.T) {

	metadata := helper.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := &helper.MockEC2Service{
		DescribeTagsFunc: helper.DescribeVolumeTagsForInstance("id-98765",
			helper.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build()),
	}

	var underTest = shared.NewEC2Instance(metadata, mockEC2Service)

	if volumes, err := underTest.AllocatedVolumes(); err != nil {
		t.Errorf("Shouldn't have failed : got error %s", err.Error())
	} else {
		if len(volumes) != 2 {
			t.Errorf("Should have got 2 allocated volumes, but got %d", len(volumes))
		}

		assertVolumesEqual(t, volumes[0], shared.NewAllocatedVolume("vol-1234567", "/dev/sda", "id-98765", nil))
		assertVolumesEqual(t, volumes[1], shared.NewAllocatedVolume("vol-54321", "/dev/sdb", "id-98765", nil))

	}

}

func assertVolumesEqual(t *testing.T, left *shared.AllocatedVolume, right *shared.AllocatedVolume) {

	if left.DeviceName != right.DeviceName || left.InstanceId != right.InstanceId || left.VolumeId != right.VolumeId {
		t.Errorf("Expected %s but got %s", left.String(), right.String())
	}
}

