package shared

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared/internal/helper"
)

func TestDetachVolumeWhenAttached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	detachVolumeFuncCalled := false
	waitUntilVolumeAvailableFuncCalled := false

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
			detachVolumeFuncCalled = true
			return helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID)(input)
		},
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			waitUntilVolumeAvailableFuncCalled = true
			return helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID)(input)
		},
	}

	saved := doAttached
	defer func() {
		doAttached = saved
	}()

	doAttached = func(volume *AllocatedVolume) (bool, error) {
		return true, nil
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

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

func TestDetachVolumeWhenDetached(t *testing.T) {

	expectedVolumeID := "vol-54321"

	detachVolumeFuncCalled := false
	waitUntilVolumeAvailableFuncCalled := false

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
			detachVolumeFuncCalled = true
			return helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID)(input)
		},
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			waitUntilVolumeAvailableFuncCalled = true
			return helper.WaitUntilVolumeAvailableForVolumeIDSuccess(expectedVolumeID)(input)
		},
	}

	saved := doAttached
	defer func() {
		doAttached = saved
	}()

	doAttached = func(volume *AllocatedVolume) (bool, error) {
		return false, nil
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	underTest.Detach()

	if detachVolumeFuncCalled || waitUntilVolumeAvailableFuncCalled {
		t.Error("No EC2 API functions should have been called")
	}

}

func TestDetachVolumeErrorCallingWaitUntilVolumeAvailableAPI(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: helper.DetachVolumeForVolumeIDSuccess(expectedVolumeID),
		WaitUntilVolumeAvailableFunc: func(input *ec2.DescribeVolumesInput) error {
			return errors.New("whoops")
		},
	}

	saved := doAttached
	defer func() {
		doAttached = saved
	}()

	doAttached = func(volume *AllocatedVolume) (bool, error) {
		return true, nil
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err == nil {
		t.Error("Detaching the volume should have failed")
	}

}

func TestDetachVolumeErrorCallingDetachVolumeAPI(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &helper.MockEC2Service{
		DetachVolumeFunc: func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
			return nil, errors.New("whoops")
		},
	}

	saved := doAttached
	defer func() {
		doAttached = saved
	}()

	doAttached = func(volume *AllocatedVolume) (bool, error) {
		return true, nil
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Detach()

	if err == nil {
		t.Error("Detaching the volume should have failed")
	}

}
