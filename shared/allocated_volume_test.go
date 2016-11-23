package shared_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
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

func TestInfo(t *testing.T) {

	expectedVolumeID := "vol-54321"
	expectedState := "blooming"
	expectedDeviceName := "/dev/sdg"

	volume := helper.NewVolumeBuilder().SetState(&expectedState).Build()

	mockEC2Service := &helper.MockEC2Service{
		DescribeVolumesFunc: helper.DescribeVolumeForID(
			expectedVolumeID,
			&ec2.DescribeVolumesOutput{
				Volumes: []*ec2.Volume{volume},
			}),
	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, expectedDeviceName, "i-11223344", mockEC2Service)

	buf := new(bytes.Buffer)
	err := underTest.Info(buf)

	if err != nil {
		t.Errorf("Attaching the volume shouldn't have failed, but I got %v", err)
	}

	infoString := buf.String()

	if !strings.Contains(infoString, expectedVolumeID) {
		t.Errorf("Info message should have contained volume id '%s', but message was : '%s'", expectedVolumeID, infoString)
	}
	if !strings.Contains(infoString, expectedState) {
		t.Errorf("Info message should have contained state '%s', but message was : '%s'", expectedState, infoString)
	}
	if !strings.Contains(infoString, expectedDeviceName) {
		t.Errorf("Info message should have contained device name '%s', but message was : '%s'", expectedDeviceName, infoString)
	}

}
