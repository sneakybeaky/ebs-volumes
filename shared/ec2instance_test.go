package shared

import (
	"testing"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
)

type mockMetadata struct {
    Metadata
	instanceId string
	region string
}

func (m *mockMetadata) InstanceID() (string, error) {
    return m.instanceId, nil
}

func (m *mockMetadata) Region() (string, error) {
	return m.region, nil
}


type mockEC2Service struct {
	ec2iface.EC2API
	expectedResourceId string
}

func (svc *mockEC2Service) DescribeTags(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {

	if *(input.Filters[0].Name) == "resource-id" {

		resourceId := input.Filters[0].Values[0]

		if *resourceId == svc.expectedResourceId {

			return &ec2.DescribeTagsOutput{
				Tags: []* ec2.TagDescription{
					&ec2.TagDescription{
						Key: aws.String("volume_/dev/sda"),
						ResourceId: resourceId,
						ResourceType: aws.String("instance"),
						Value: aws.String("vol-1234567"),
					},
					&ec2.TagDescription{
						Key: aws.String("volume_/dev/sdb"),
						ResourceId: resourceId,
						ResourceType: aws.String("instance"),
						Value: aws.String("vol-54321"),
					},
				},
			}, nil
		}

		return nil, fmt.Errorf("No tags for %s",*resourceId)

	}

	return nil, errors.New("No resource id set")
}

func TestFindAllocatedVolumes(t *testing.T) {

	metadata := &mockMetadata{instanceId: "id-98765", region: "erewhon"}

	var underTest = NewEC2Instance(metadata,&mockEC2Service{expectedResourceId:"id-98765"})

	if volumes, err := underTest.AllocatedVolumes(); err != nil {
		t.Errorf("Shouldn't have failed : got error %s",err.Error())
	} else {
		if len(volumes) != 2 {
			t.Errorf("Should have got 2 allocated volumes, but got %d", len(volumes))
		}

		assertVolumesEqual(t, volumes[0],NewAllocatedVolume("vol-1234567","/dev/sda","id-98765",nil))
		assertVolumesEqual(t, volumes[1],NewAllocatedVolume("vol-54321","/dev/sdb","id-98765",nil))

	}

}

func assertVolumesEqual(t *testing.T, left *AllocatedVolume, right *AllocatedVolume) {

	if (left.DeviceName != right.DeviceName || left.InstanceId != right.InstanceId || left.VolumeId != right.VolumeId) {
		t.Errorf("Expected %s but got %s",left.String(),right.String())
	}
}