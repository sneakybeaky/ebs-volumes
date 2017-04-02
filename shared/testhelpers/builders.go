package testhelpers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// DescribeTagsOutputBuilder helps construct an ec2.DescribeTagsOutput structure for humans
type DescribeTagsOutputBuilder struct {
	tagDescriptions []*ec2.TagDescription
}

// NewDescribeTagsOutputBuilder returns a new DescribeTagsOutputBuilder
func NewDescribeTagsOutputBuilder() *DescribeTagsOutputBuilder {
	return &DescribeTagsOutputBuilder{}
}

// WithVolume adds an allocated volume tag
func (builder DescribeTagsOutputBuilder) WithVolume(DeviceName string, InstanceID string, VolumeID string) DescribeTagsOutputBuilder {
	builder.tagDescriptions = append(builder.tagDescriptions, &ec2.TagDescription{
		Key:          aws.String(fmt.Sprintf("volume_%s", DeviceName)),
		ResourceId:   aws.String(InstanceID),
		ResourceType: aws.String("instance"),
		Value:        aws.String(VolumeID),
	})

	return builder
}

// DetachVolumes sets the tag to indicate volumes should be detached
func (builder DescribeTagsOutputBuilder) DetachVolumes(instanceID string) DescribeTagsOutputBuilder {
	return builder.DetachVolumesValue(instanceID, "true")
}

// DetachVolumesValue sets the value for the flag for detaching volumes
func (builder DescribeTagsOutputBuilder) DetachVolumesValue(instanceID string, value string) DescribeTagsOutputBuilder {
	builder.tagDescriptions = append(builder.tagDescriptions, &ec2.TagDescription{
		Key:          aws.String("detach_volumes"),
		ResourceId:   aws.String(instanceID),
		ResourceType: aws.String("instance"),
		Value:        aws.String(value),
	})

	return builder
}

// Build generates the final ec2.DescribeTagsOutput structure
func (builder DescribeTagsOutputBuilder) Build() *ec2.DescribeTagsOutput {
	return &ec2.DescribeTagsOutput{
		Tags: builder.tagDescriptions,
	}
}

// VolumeBuilder helps construct an ec2.Volume structure for humans
type VolumeBuilder struct {
	state *string
}

// NewVolumeBuilder returns a new VolumeBuilder
func NewVolumeBuilder() *VolumeBuilder {
	return &VolumeBuilder{}
}

// SetState sets the state value
func (builder VolumeBuilder) SetState(state *string) VolumeBuilder {
	builder.state = state
	return builder
}

// Build returns a populated Volume structure
func (builder VolumeBuilder) Build() *ec2.Volume {
	return &ec2.Volume{State: builder.state}
}
