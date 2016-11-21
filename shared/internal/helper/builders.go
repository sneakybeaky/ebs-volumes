package helper

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type DescribeTagsOutputBuilder struct {
	TagDescriptions []*ec2.TagDescription
}

func NewDescribeTagsOutputBuilder() *DescribeTagsOutputBuilder {
	return &DescribeTagsOutputBuilder{}
}

func (builder DescribeTagsOutputBuilder) WithVolume(DeviceName string, InstanceId string, VolumeID string) DescribeTagsOutputBuilder {
	builder.TagDescriptions = append(builder.TagDescriptions, &ec2.TagDescription{
		Key:          aws.String(fmt.Sprintf("volume_%s", DeviceName)),
		ResourceId:   aws.String(InstanceId),
		ResourceType: aws.String("instance"),
		Value:        aws.String(VolumeID),
	})

	return builder
}

func (builder DescribeTagsOutputBuilder) Build() *ec2.DescribeTagsOutput {
	return &ec2.DescribeTagsOutput{
		Tags: builder.TagDescriptions,
	}
}
