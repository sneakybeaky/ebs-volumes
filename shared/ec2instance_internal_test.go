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

	attached := captureAttachedVolumes(expectedVolumes, underTest)

	checkExpectedVolumesWereAttached(expectedVolumes, attached, t)

}

func TestDetachAllocatedVolumes(t *testing.T) {

	instanceID := "id-98765"
	metadata := helper.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := helper.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = helper.DescribeVolumeTagsForInstance(instanceID,
		helper.NewDescribeTagsOutputBuilder().DetachVolumes(instanceID).WithVolume("/dev/sda", instanceID, "vol-1234567").WithVolume("/dev/sdb", instanceID, "vol-54321").Build())

	expectedVolumes := []string{"vol-1234567", "vol-54321"}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	detached := captureDetachedVolumes(expectedVolumes, underTest)

	checkExpectedVolumesWereDetached(expectedVolumes, detached, t)

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

func checkExpectedVolumesWereAttached(expectedVolumes []string, attached map[string]bool, t *testing.T) {
	for _, expectedVolume := range expectedVolumes {
		if _, attached := attached[expectedVolume]; !attached {
			t.Errorf("Volume %s should have been attached, but wasn't", expectedVolume)
		}
	}
}

func checkExpectedVolumesWereDetached(expectedVolumes []string, detached map[string]bool, t *testing.T) {
	for _, expectedVolume := range expectedVolumes {
		if _, detached := detached[expectedVolume]; !detached {
			t.Errorf("Volume %s should have been detached, but wasn't", expectedVolume)
		}
	}
}

func captureAttachedVolumes(expectedVolumes []string, underTest *EC2Instance) map[string]bool {

	saved := attachVolume
	defer func() {
		attachVolume = saved
	}()

	attachedChannel := make(chan string, len(expectedVolumes))
	attachVolume = func(volume *AllocatedVolume) {
		attachedChannel <- volume.VolumeId
	}
	underTest.AttachVolumes()
	close(attachedChannel)
	attached := make(map[string]bool)
	for volumeId := range attachedChannel {
		attached[volumeId] = true
	}
	return attached
}

func captureDetachedVolumes(expectedVolumes []string, underTest *EC2Instance) map[string]bool {

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachedChannel := make(chan string, len(expectedVolumes))
	detachVolume = func(volume *AllocatedVolume) {
		detachedChannel <- volume.VolumeId
	}
	underTest.DetachVolumes()
	close(detachedChannel)
	detached := make(map[string]bool)
	for volumeId := range detachedChannel {
		detached[volumeId] = true
	}
	return detached
}
