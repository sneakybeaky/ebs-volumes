package shared

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"testing"
)

type mockMetadata struct {
	Metadata
	instanceId string
	region     string
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
	DescribeTagsFunc   func(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error)
}

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

func (svc *mockEC2Service) DescribeTags(input *ec2.DescribeTagsInput) (*ec2.DescribeTagsOutput, error) {

	expectedOutputBuilder := NewDescribeTagsOutputBuilder().WithVolume("/dev/sda", svc.expectedResourceId, "vol-1234567").WithVolume("/dev/sdb", svc.expectedResourceId, "vol-54321")

	if *(input.Filters[0].Name) == "resource-id" {

		resourceId := input.Filters[0].Values[0]

		if *resourceId == svc.expectedResourceId {

			return expectedOutputBuilder.Build(), nil
		}

		return nil, fmt.Errorf("No tags for %s", *resourceId)

	}

	return nil, errors.New("No resource id set")
}

func TestFindAllocatedVolumes(t *testing.T) {

	metadata := &mockMetadata{instanceId: "id-98765", region: "erewhon"}

	var underTest = NewEC2Instance(metadata, &mockEC2Service{expectedResourceId: "id-98765"})

	if volumes, err := underTest.AllocatedVolumes(); err != nil {
		t.Errorf("Shouldn't have failed : got error %s", err.Error())
	} else {
		if len(volumes) != 2 {
			t.Errorf("Should have got 2 allocated volumes, but got %d", len(volumes))
		}

		assertVolumesEqual(t, volumes[0], NewAllocatedVolume("vol-1234567", "/dev/sda", "id-98765", nil))
		assertVolumesEqual(t, volumes[1], NewAllocatedVolume("vol-54321", "/dev/sdb", "id-98765", nil))

	}

}

func assertVolumesEqual(t *testing.T, left *AllocatedVolume, right *AllocatedVolume) {

	if left.DeviceName != right.DeviceName || left.InstanceId != right.InstanceId || left.VolumeId != right.VolumeId {
		t.Errorf("Expected %s but got %s", left.String(), right.String())
	}
}

func TestAttachAllocatedVolumes(t *testing.T) {

	metadata := &mockMetadata{instanceId: "id-98765", region: "erewhon"}

	var underTest = NewEC2Instance(metadata, &mockEC2Service{expectedResourceId: "id-98765"})

	saved := attachVolume
	defer func() { attachVolume = saved }()

	set := make(map[string]struct{}, 2)
	attachVolume = func(volume *AllocatedVolume) { set[volume.VolumeId] = struct{}{} }

	underTest.AttachVolumes()

	if len(set) != 2 {
		t.Errorf("Should have been 2 volumes attached, but %d were", len(set))
	}

	expectedVolumes := []string{"vol-1234567", "vol-54321"}
	for _, expectedVolume := range expectedVolumes {
		if _, ok := set[expectedVolume]; !ok {
			t.Errorf("Volume %s should have been attached, but wasn't", expectedVolume)
		}
	}

}
