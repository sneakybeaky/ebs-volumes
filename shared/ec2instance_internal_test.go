package shared

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestAttachAllocatedVolumes(t *testing.T) {

	metadata := helper.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance("id-98765",
		helper.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build())

	expectedVolumes := []string{"vol-1234567", "vol-54321"}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := attachVolume
	defer func() {
		attachVolume = saved
	}()

	attachedVolumes := ""
	attachVolume = func(volume *AllocatedVolume) {
		attachedVolumes += fmt.Sprintf("%s:", volume.VolumeId)
	}

	underTest.AttachVolumes()

	for _, expectedVolume := range expectedVolumes {
		if attached := strings.Contains(attachedVolumes, expectedVolume); !attached {
			t.Errorf("Volume %s should have been attached, but wasn't in the attached volumes %v ", expectedVolume, strings.Split(attachedVolumes, ":"))
		}
	}

}
