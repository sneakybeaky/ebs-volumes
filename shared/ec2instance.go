package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

type AllocatedVolume struct {
	VolumeID string
	Device   string
}

const volume_tag_prefix = "volume_"

// A EC2InstanceMetadata provides metadata about an EC2 instance.
type EC2Instance struct {
	EC2      *ec2.EC2
	metadata *EC2InstanceMetadata
}

func NewEC2Instance(metadata *EC2InstanceMetadata, session *session.Session, cfg ...*aws.Config) *EC2Instance {

	return &EC2Instance{
		EC2: ec2.New(session, cfg...),
		metadata: metadata,
	}

}

func (e EC2Instance) AttachedVolumes() ([]*ec2.Volume, error) {

	instanceid, err := e.metadata.InstanceID()

	if err != nil {
		return nil, err
	}

	params := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.instance-id"),
				Values: []*string{
					aws.String(instanceid),
				},
			},
		},
	}

	resp, err := e.EC2.DescribeVolumes(params)

	if err != nil {
		return nil, err
	}

	return resp.Volumes, nil

}

func (e EC2Instance) Tags() ([]*ec2.TagDescription, error) {

	instanceid, err := e.metadata.InstanceID()

	if err != nil {
		return nil, err
	}

	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{// Required
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(instanceid), // Required
				},
			},
		},
		MaxResults: aws.Int64(1000),
	}
	resp, err := e.EC2.DescribeTags(params)

	if err != nil {
		return nil, err
	}

	return resp.Tags, nil
}

func (e EC2Instance) AllocatedVolumes() ([]AllocatedVolume, error) {
	var allocated []AllocatedVolume

	if tags, err := e.Tags(); err != nil {
		return allocated, err
	} else {
		for _, tag := range tags {
			if strings.HasPrefix(*tag.Key, volume_tag_prefix) {

				key := *tag.Key
				device := key[len(volume_tag_prefix):]
				allocated = append(allocated, AllocatedVolume{VolumeID:*tag.Value, Device: device})
			}
		}
	}

	return allocated, nil
}