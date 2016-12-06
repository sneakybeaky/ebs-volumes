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

func TestDetachAllocatedVolumes(t *testing.T) {

	instanceID := "id-98765"
	metadata := helper.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance(instanceID,
		helper.NewDescribeTagsOutputBuilder().DetachVolumes(instanceID).WithVolume("/dev/sda", instanceID, "vol-1234567").WithVolume("/dev/sdb", instanceID, "vol-54321").Build())

	expectedVolumes := []string{"vol-1234567", "vol-54321"}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachedVolumes := ""
	detachVolume = func(volume *AllocatedVolume) {
		detachedVolumes += fmt.Sprintf("%s:", volume.VolumeId)
	}

	underTest.DetachVolumes()

	for _, expectedVolume := range expectedVolumes {
		if attached := strings.Contains(detachedVolumes, expectedVolume); !attached {
			t.Errorf("Volume %s should have been detached, but wasn't in the detached volumes %v ", expectedVolume, strings.Split(detachedVolumes, ":"))
		}
	}

}

func TestVolumesNotDetachedWhenTagUnset(t *testing.T) {

	metadata := helper.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance("id-98765",
		helper.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachedVolumes := ""
	detachVolume = func(volume *AllocatedVolume) {
		detachedVolumes += fmt.Sprintf("%s:", volume.VolumeId)
	}

	underTest.DetachVolumes()

	if detachedVolumes != "" {
		t.Errorf("No volumes should have been detached, but %s were ", strings.Split(detachedVolumes, ":"))
	}

}

func TestVolumesNotDetachedWhenTagValueIsNotTrue(t *testing.T) {

	instanceID := "id-98765"

	metadata := helper.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance(instanceID,
		helper.NewDescribeTagsOutputBuilder().DetachVolumesValue(instanceID, "false").WithVolume("/dev/sda", instanceID, "vol-1234567").WithVolume("/dev/sdb", instanceID, "vol-54321").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachedVolumes := ""
	detachVolume = func(volume *AllocatedVolume) {
		detachedVolumes += fmt.Sprintf("%s:", volume.VolumeId)
	}

	underTest.DetachVolumes()

	if detachedVolumes != "" {
		t.Errorf("No volumes should have been detached, but %s were ", strings.Split(detachedVolumes, ":"))
	}

}
