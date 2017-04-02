package shared

import (
	"bytes"
	"strings"
	"testing"

	"errors"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/ebs-volumes/shared/testhelpers"
)

func TestInfo(t *testing.T) {

	expectedVolumeID := "vol-54321"
	expectedState := "blooming"
	expectedDeviceName := "/dev/sdg"

	volume := testhelpers.NewVolumeBuilder().SetState(&expectedState).Build()

	mockEC2Service := &testhelpers.MockEC2Service{
		DescribeVolumesFunc: testhelpers.DescribeVolumeForID(
			expectedVolumeID,
			&ec2.DescribeVolumesOutput{
				Volumes: []*ec2.Volume{volume},
			}),
	}

	underTest := NewAllocatedVolume(expectedVolumeID, expectedDeviceName, "i-11223344", mockEC2Service)

	buf := new(bytes.Buffer)
	err := underTest.Info(buf)

	if err != nil {
		t.Errorf("Getting info shouldn't have failed, but I got %v", err)
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

func TestInfoErrorCallingDescribeVolumesAPI(t *testing.T) {

	expectedVolumeID := "vol-54321"

	mockEC2Service := &testhelpers.MockEC2Service{
		DescribeVolumesFunc: func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
			return nil, errors.New("whoops")
		},
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)

	err := underTest.Info(ioutil.Discard)

	if err == nil {
		t.Error("Getting the volume info should have returned an error")
	}

}

func TestAttachedStatusWhenDetached(t *testing.T) {
	expectedVolumeID := "vol-54321"

	mockEC2Service := &testhelpers.MockEC2Service{
		DescribeVolumesFunc: testhelpers.DescribeVolumeForID(
			expectedVolumeID,
			&ec2.DescribeVolumesOutput{
				Volumes: []*ec2.Volume{},
			}),
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)
	attached, _ := underTest.Attached()

	if attached != false {
		t.Error("The volume is not attached")
	}
}

func TestAttachedStatusWhenAttached(t *testing.T) {
	expectedVolumeID := "vol-54321"

	volume := testhelpers.NewVolumeBuilder().Build()

	mockEC2Service := &testhelpers.MockEC2Service{
		DescribeVolumesFunc: testhelpers.DescribeVolumeForID(
			expectedVolumeID,
			&ec2.DescribeVolumesOutput{
				Volumes: []*ec2.Volume{volume},
			}),
	}

	underTest := NewAllocatedVolume(expectedVolumeID, "/dev/sdg", "i-11223344", mockEC2Service)
	attached, _ := underTest.Attached()

	if attached != true {
		t.Error("The volume is attached")
	}
}
