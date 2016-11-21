package shared

import (
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestAttachAllocatedVolumes(t *testing.T) {

	metadata := helper.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance("id-98765",
		helper.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := attachVolume
	defer func() {
		attachVolume = saved
	}()

	set := make(map[string]struct{}, 2)
	attachVolume = func(volume *AllocatedVolume) {
		set[volume.VolumeId] = struct{}{}
	}

	underTest.AttachVolumes()

	if len(set) != 2 {
		t.Errorf("Should have been 2 volumes attached, but %d were", len(set))
	}

	expectedVolumes := []string{"vol-1234567", "vol-54321"}
	for _, expectedVolume := range expectedVolumes {
		if _, ok := set[expectedVolume]; !ok {
			t.Errorf("Volume %s should have been attached, but wasn't", expectedVolume)
		}
	}

}
