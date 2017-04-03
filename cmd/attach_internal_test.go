package cmd

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/ebs-volumes/shared"
	"github.com/sneakybeaky/ebs-volumes/shared/testhelpers"
)

func TestErrorReturnedWhenErrorDuringAttach(t *testing.T) {
	instanceID := "id-98765"
	metadata := testhelpers.NewMockMetadata(instanceID, "erewhon")

	mockEC2Service := testhelpers.NewMockEC2Service()

	mockEC2Service.DescribeTagsFunc = testhelpers.DescribeVolumeTagsForInstance(instanceID,
		testhelpers.NewDescribeTagsOutputBuilder().DetachVolumes(instanceID).WithVolume("/dev/sda", instanceID, "vol-1234567").Build())

	mockEC2Service.DescribeVolumesFunc = func(*ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
		return nil, errors.New("Whoops")
	}

	saved := getInstance
	defer func() {
		getInstance = saved
	}()

	getInstance = func() (*shared.EC2Instance, error) {
		return shared.NewEC2Instance(metadata, mockEC2Service), nil
	}

	err := attachCmd.Execute()

	if err == nil {
		t.Error("No error returned")
	}
}
