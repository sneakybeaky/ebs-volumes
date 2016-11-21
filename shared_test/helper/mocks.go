package helper

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/aws-volumes/shared/iface"
)

type MockMetadata struct {
	iface.Metadata
	instanceId string
	region     string
}

func NewMockMetadata(instanceId string, region string) *MockMetadata {
	return &MockMetadata{instanceId: instanceId, region: region}
}

func (m *MockMetadata) InstanceID() (string, error) {
	return m.instanceId, nil
}

func (m *MockMetadata) Region() (string, error) {
	return m.region, nil
}

type MockEC2Service struct {
	ec2iface.EC2API
	DescribeTagsFunc func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error)
}

func (svc *MockEC2Service) DescribeTags(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	return svc.DescribeTagsFunc(input)
}

func DescribeVolumeTagsForInstance(instanceId string, output *ec2.DescribeTagsOutput) func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	return func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
		if *(input.Filters[0].Name) == "resource-id" {

			resourceId := input.Filters[0].Values[0]

			if *resourceId == instanceId {

				return output, nil
			}

			return nil, fmt.Errorf("No tags for %s", *resourceId)

		}

		return nil, errors.New("No resource id set")
	}
}
