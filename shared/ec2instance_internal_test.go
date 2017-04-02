package shared

import (
	"testing"

	"errors"

	"github.com/sneakybeaky/ebs-volumes/shared/testhelpers"
)

func TestAttachAllocatedVolumes(t *testing.T) {

	metadata := testhelpers.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance("id-98765",
		testhelpers.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build())

	expectedVolumes := []string{"vol-1234567", "vol-54321"}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	attached := captureAttachedVolumes(expectedVolumes, underTest)

	checkExpectedVolumesWereAttached(expectedVolumes, attached, t)

}

func TestDetachAllocatedVolumes(t *testing.T) {

	instanceID := "id-98765"
	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().DetachVolumes(instanceID).WithVolume("/dev/sda", instanceID, "vol-1234567").WithVolume("/dev/sdb", instanceID, "vol-54321").Build())

	expectedVolumes := []string{"vol-1234567", "vol-54321"}

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	detached := captureDetachedVolumes(expectedVolumes, underTest)

	checkExpectedVolumesWereDetached(expectedVolumes, detached, t)

}

func TestVolumesNotDetachedWhenTagUnset(t *testing.T) {

	metadata := testhelpers.NewMockMetadata("id-98765", "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance("id-98765",
		testhelpers.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", "id-98765", "vol-1234567").WithVolume("/dev/sdb", "id-98765", "vol-54321").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	detached := captureDetachedVolumes([]string{}, underTest)

	if len(detached) > 0 {
		t.Errorf("No volumes should have been detached, but %d were ", len(detached))
	}

}

func TestVolumesNotDetachedWhenTagValueIsNotTrue(t *testing.T) {

	instanceID := "id-98765"

	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().DetachVolumesValue(instanceID, "false").WithVolume("/dev/sda", instanceID, "vol-1234567").WithVolume("/dev/sdb", instanceID, "vol-54321").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	detached := captureDetachedVolumes([]string{}, underTest)

	if len(detached) > 0 {
		t.Errorf("No volumes should have been detached, but %d were ", len(detached))
	}

}

func TestErrorReturnedWhenDetachVolumeErrors(t *testing.T) {
	instanceID := "id-98765"
	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().DetachVolumes(instanceID).WithVolume("/dev/sda", instanceID, "vol-1234567").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachVolume = func(volume *AllocatedVolume) error {
		return errors.New("Couldn't detach")
	}
	error := underTest.DetachVolumes()

	if error == nil {
		t.Error("Error should have been returned")
	}

}

func TestErrorReturnedWhenAttachVolumeErrors(t *testing.T) {
	instanceID := "id-98765"
	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", instanceID, "vol-1234567").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := attachVolume
	defer func() {
		attachVolume = saved
	}()

	attachVolume = func(volume *AllocatedVolume) error {
		return errors.New("Couldn't attach")
	}
	error := underTest.AttachVolumes()

	if error == nil {
		t.Error("Error should have been returned")
	}

}

func TestErrorReturnedWhenInfoErrors(t *testing.T) {
	instanceID := "id-98765"
	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", instanceID, "vol-1234567").Build())

	var underTest = NewEC2Instance(metadata, mockEC2Service)

	saved := showVolumeInfo
	defer func() {
		showVolumeInfo = saved
	}()

	showVolumeInfo = func(volume *AllocatedVolume) error {
		return errors.New("Couldn't attach")
	}
	error := underTest.ShowVolumesInfo()

	if error == nil {
		t.Error("Error should have been returned")
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
	attachVolume = func(volume *AllocatedVolume) error {
		attachedChannel <- volume.VolumeID
		return nil
	}
	underTest.AttachVolumes()
	close(attachedChannel)
	attached := make(map[string]bool)
	for volumeID := range attachedChannel {
		attached[volumeID] = true
	}
	return attached
}

func captureDetachedVolumes(expectedVolumes []string, underTest *EC2Instance) map[string]bool {

	saved := detachVolume
	defer func() {
		detachVolume = saved
	}()

	detachedChannel := make(chan string, len(expectedVolumes))
	detachVolume = func(volume *AllocatedVolume) error {
		detachedChannel <- volume.VolumeID
		return nil
	}
	underTest.DetachVolumes()
	close(detachedChannel)
	detached := make(map[string]bool)
	for volumeID := range detachedChannel {
		detached[volumeID] = true
	}
	return detached
}
