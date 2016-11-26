package shared_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared"
	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
	"errors"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	detachVolumeFuncCalled := false
	waitUntilVolumeAvailableFuncCalled := false

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc:    func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
			detachVolumeFuncCalled = true
			return helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID)(input)
		},
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			waitUntilVolumeAvailableFuncCalled = true
			return helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID)(input)
		},

	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err != nil {
		t.Errorf("Detaching the volume shouldn't have failed, but I got %v", err)
	}

	if !detachVolumeFuncCalled {
		t.Error("The AWS API DetachVolume function wasn't called ")
	}

	if !waitUntilVolumeAvailableFuncCalled {
		t.Error("The AWS API WaitUntilVolumeAvailable function wasn't called ")
	}
}

func TestAttachVolumeWhenDetached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	waitUntilVolumeAvailableFuncCalled := false
	attachVolumeFuncCalled := false
	waitUntilVolumeInUseFunc := false

	mockEC2Service := &helper.MockEC2Service{
		AttachVolumeFunc: func(input *ec2.AttachVolumeInput) (*ec2.VolumeAttachment, error) {
			attachVolumeFuncCalled = true
			return helper.AttachVolumeForVolumeIDSuccess(expectedVolumeID)(input)
		},
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			waitUntilVolumeAvailableFuncCalled = true
			return helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID)(input)
		},
		WaitUntilVolumeInUseFunc: func(input *ec2.DescribeVolumesInput) error {
			waitUntilVolumeInUseFunc = true
			return helper.WaitUntilVolumeInUseForVolumeIDSuccess(expectedVolumeID)(input)
		},
	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Attach()

	if err != nil {
		t.Errorf("Attaching the volume shouldn't have failed, but I got %v", err)
	}

	if !waitUntilVolumeAvailableFuncCalled {
		t.Error("The AWS API WaitUntilVolumeAvailable function wasn't called ")
	}

	if !attachVolumeFuncCalled {
		t.Error("The AWS API AttachVolume function wasn't called ")
	}

	if !waitUntilVolumeInUseFunc {
		t.Error("The AWS API WaitUntilVolumeInUseFunc function wasn't called ")
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

func TestDetachVolumeErrorCallingDetachVolumeAPI(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc:    func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
			return nil,errors.New("whoops")
		},

	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err == nil {
		t.Error("Detaching the volume should have failed")
	}

}

func TestDetachVolumeErrorCallingWaitUntilVolumeAvailableAPI(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc:     helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			return errors.New("whoops")
		},

	}

	underTest := shared.NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err == nil {
		t.Error("Detaching the volume should have failed")
	}

}
