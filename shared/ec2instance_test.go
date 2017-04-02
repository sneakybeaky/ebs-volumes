package shared

import (
	"testing"

	"github.com/sneakybeaky/ebs-volumes/shared/testhelpers"
)

func TestFindAllocatedVolumes(t *testing.T) {

	metadata := testhelpers.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := &testhelpers.MockEC2Service{
		DescribeTagsFunc: testhelpers.DescribeVolumeTagsForInstance("id-98765",
			testhelpers.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build()),
	}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	if volumes, err := underTest.AllocatedVolumes(); err != nil {
		t.Errorf("Shouldn't have failed : got error %s", err.Error())
	} else {
		if len(volumes) != 2 {
			t.Errorf("Should have got 2 allocated volumes, but got %d", len(volumes))
		}

		assertVolumesEqual(t, volumes[0], NewAllocatedVolume("vol-1234567", "/dev/sda", "id-98765", nil))
		assertVolumesEqual(t, volumes[1], NewAllocatedVolume("vol-54321", "/dev/sdb", "id-98765", nil))

	}

}

func assertVolumesEqual(t *testing.T, left *AllocatedVolume, right *AllocatedVolume) {

	if left.DeviceName != right.DeviceName || left.InstanceID != right.InstanceID || left.VolumeID != right.VolumeID {
		t.Errorf("Expected %s but got %s", left.String(), right.String())
	}
}
