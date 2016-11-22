package helper

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/aws-volumes/shared/iface"
)

// MockMetadata enables plugable behaviour for testing
type MockMetadata struct {
	iface.Metadata
	instanceID string
	region     string
}

// NewMockMetadata returns a new MockMetadata instance
func NewMockMetadata(instanceID string, region string) *MockMetadata {
	return &MockMetadata{instanceID: instanceID, region: region}
}

// InstanceID returns the instance id
func (m *MockMetadata) InstanceID() (string, error) {
	return m.instanceID, nil
}

// Region returns the instance id
func (m *MockMetadata) Region() (string, error) {
	return m.region, nil
}

// MockEC2Service enables plugable behaviour for testing
type MockEC2Service struct {
	ec2iface.EC2API
	DescribeTagsFunc func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error)
}

// NewMockEC2Service returns a new instance of NewMockEC2Service
func NewMockEC2Service() *MockEC2Service {
	return &MockEC2Service{}
}

// DescribeTags pass through that calls the DescribeTagsFunc on the mock
func (svc *MockEC2Service) DescribeTags(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	return svc.DescribeTagsFunc(input)
}

//DescribeVolumeTagsForInstance returns a function that returns a canned response for a given instanceId
func DescribeVolumeTagsForInstance(instanceID string, output *ec2.DescribeTagsOutput) func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	return func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
		if *(input.Filters[0].Name) == "resource-id" {

			resourceID := input.Filters[0].Values[0]

			if *resourceID == instanceID {

				return output, nil
			}

			return nil, fmt.Errorf("No tags for %s", *resourceID)

		}

		return nil, errors.New("No resource id set")
	}
}
