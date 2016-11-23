package helper

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/request"
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
	AttachVolumeFunc             func(*ec2.AttachVolumeInput) (*ec2.VolumeAttachment, error)
	DescribeTagsFunc             func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error)
	DetachVolumeFunc             func(*ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error)
	DescribeVolumesRequestFunc   func(*ec2.DescribeVolumesInput) (*request.Request, *ec2.DescribeVolumesOutput)
	WaitUntilVolumeAvailableFunc func(*ec2.DescribeVolumesInput) error
	WaitUntilVolumeInUseFunc     func(*ec2.DescribeVolumesInput) error
	DescribeVolumesFunc          func(*ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error)
}

// NewMockEC2Service returns a new instance of NewMockEC2Service
func NewMockEC2Service() *MockEC2Service {
	return &MockEC2Service{}
}

// AttachVolume pass through that calls the AttachVolumeFunc on the mock
func (svc *MockEC2Service) AttachVolume(input *ec2.AttachVolumeInput) (*ec2.VolumeAttachment, error) {
	return svc.AttachVolumeFunc(input)
}

// DescribeTags pass through that calls the DescribeTagsFunc on the mock
func (svc *MockEC2Service) DescribeTags(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {
	return svc.DescribeTagsFunc(input)
}

// DetachVolume pass through that calls the DetachVolumeFunc on the mock
func (svc *MockEC2Service) DetachVolume(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
	return svc.DetachVolumeFunc(input)
}

// DescribeVolumesRequest pass through that calls the DescribeVolumesRequestFunc on the mock
func (svc *MockEC2Service) DescribeVolumesRequest(input *ec2.DescribeVolumesInput) (*request.Request, *ec2.DescribeVolumesOutput) {
	return svc.DescribeVolumesRequestFunc(input)
}

// WaitUntilVolumeAvailable pass through that calls the WaitUntilVolumeAvailableFunc on the mock
func (svc *MockEC2Service) WaitUntilVolumeAvailable(input *ec2.DescribeVolumesInput) error {
	return svc.WaitUntilVolumeAvailableFunc(input)
}

// WaitUntilVolumeInUse pass through that calls the WaitUntilVolumeAvailableFunc on the mock
func (svc *MockEC2Service) WaitUntilVolumeInUse(input *ec2.DescribeVolumesInput) error {
	return svc.WaitUntilVolumeInUseFunc(input)
}

// DescribeVolumes pass through that calls the DescribeVolumesFunc on the mock
func (svc *MockEC2Service) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return svc.DescribeVolumesFunc(input)
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

//DetachVolumeForVolumeIDSuccess returns an empty structure for a specific VolumeID. All others return an error
func DetachVolumeForVolumeIDSuccess(volumeID string) func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
	return func(input *ec2.DetachVolumeInput) (*ec2.VolumeAttachment, error) {
		if *(input.VolumeId) == volumeID {

			return &ec2.VolumeAttachment{}, nil

		}

		return nil, fmt.Errorf("Unexpected volume id %s", *input.VolumeId)
	}
}

//AttachVolumeForVolumeIDSuccess returns an empty structure for a specific VolumeID. All others return an error
func AttachVolumeForVolumeIDSuccess(volumeID string) func(input *ec2.AttachVolumeInput) (*ec2.VolumeAttachment, error) {
	return func(input *ec2.AttachVolumeInput) (*ec2.VolumeAttachment, error) {
		if *(input.VolumeId) == volumeID {

			return &ec2.VolumeAttachment{}, nil

		}

		return nil, fmt.Errorf("Unexpected volume id %s", *input.VolumeId)
	}
}

//WaitUntilVolumeAvailableForVolumeIDSuccess returns a function that returns a nil error for the supplied volume id otherwise a non nil error
func WaitUntilVolumeAvailableForVolumeIDSuccess(volumeID string) func(input *ec2.DescribeVolumesInput) error {
	return func(input *ec2.DescribeVolumesInput) error {

		if *input.VolumeIds[0] == volumeID {
			return nil
		}

		return fmt.Errorf("Unexpected volume id %s", *input.VolumeIds[0])

	}
}

//WaitUntilVolumeInUseForVolumeIDSuccess returns a function that returns a nil error for the supplied volume id otherwise a non nil error
func WaitUntilVolumeInUseForVolumeIDSuccess(volumeID string) func(input *ec2.DescribeVolumesInput) error {
	return func(input *ec2.DescribeVolumesInput) error {

		if *input.VolumeIds[0] == volumeID {
			return nil
		}

		return fmt.Errorf("Unexpected volume id %s", *input.VolumeIds[0])

	}
}

//DescribeVolumeForID returns a function that returns the supplied output for the supplied volume id otherwise a non nil error
func DescribeVolumeForID(volumeID string, output *ec2.DescribeVolumesOutput) func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {
	return func(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {

		if *input.VolumeIds[0] == volumeID {
			return output, nil
		}

		return nil, fmt.Errorf("Unexpected volume id %s", *input.VolumeIds[0])

	}
}
