package helper

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// DescribeTagsOutputBuilder helps construct an ec2.DescribeTagsOutput structure for humans
type DescribeTagsOutputBuilder struct {
	TagDescriptions []*ec2.TagDescription
}

// NewDescribeTagsOutputBuilder returns a new DescribeTagsOutputBuilder
func NewDescribeTagsOutputBuilder() *DescribeTagsOutputBuilder {
	return &DescribeTagsOutputBuilder{}
}

// WithVolume adds an allocated volume tag
func (builder DescribeTagsOutputBuilder) WithVolume(DeviceName string, InstanceID string, VolumeID string) DescribeTagsOutputBuilder {
	builder.TagDescriptions = append(builder.TagDescriptions, &ec2.TagDescription{
		Key:          aws.String(fmt.Sprintf("volume_%s", DeviceName)),
		ResourceId:   aws.String(InstanceID),
		ResourceType: aws.String("instance"),
		Value:        aws.String(VolumeID),
	})

	return builder
}

func (builder DescribeTagsOutputBuilder) DetachVolumes(instanceID string) DescribeTagsOutputBuilder {
	return builder.DetachVolumesValue(instanceID,"true")
}

func (builder DescribeTagsOutputBuilder) DetachVolumesValue(instanceID string, value string) DescribeTagsOutputBuilder {
	builder.TagDescriptions = append(builder.TagDescriptions, &ec2.TagDescription{
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
		Tags: builder.TagDescriptions,
	}
}
