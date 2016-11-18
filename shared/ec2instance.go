package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
	"github.com/sneakybeaky/aws-volumes/shared/log"
)

const volume_tag_prefix = "volume_"

// A EC2InstanceMetadata provides metadata about an EC2 instance.
type EC2Instance struct {
	EC2      *ec2.EC2
	metadata *EC2InstanceMetadata
}

func NewEC2Instance(metadata *EC2InstanceMetadata, session *session.Session, cfg ...*aws.Config) *EC2Instance {

	return &EC2Instance{
		EC2:      ec2.New(session, cfg...),
		metadata: metadata,
	}

}

func (e EC2Instance) Tags() ([]*ec2.TagDescription, error) {

	instanceid, err := e.metadata.InstanceID()

	if err != nil {
		return nil, err
	}

	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{ // Required
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

func (e EC2Instance) AllocatedVolumes() ([]*AllocatedVolume, error) {
	var allocated []*AllocatedVolume

	if tags, err := e.Tags(); err != nil {
		return allocated, err
	} else {
		for _, tag := range tags {
			if strings.HasPrefix(*tag.Key, volume_tag_prefix) {

				key := *tag.Key
				device := key[len(volume_tag_prefix):]
				allocated = append(allocated, NewAllocatedVolume(*tag.Value, device, *tag.ResourceId, e.EC2))
			}
		}
	}

	return allocated, nil
}

func (e EC2Instance) DetachVolumes() {

	if volumes, err := e.AllocatedVolumes(); err != nil {
		log.Error.Printf("Unable to find allocated volumes : %s", err)
	} else {

		done := make(chan int)
		defer close(done)

		for _, volume := range volumes {

			go func(volume *AllocatedVolume) {

				if err := volume.Detach(); err != nil {
					log.Error.Printf("Unable to detach volume : %s\n", err)
				}
				done <- 1
			}(volume)

		}

		for i := 0; i < len(volumes); i++ {
			<-done
		}
	}

}

func (e EC2Instance) AttachVolumes() {
	if volumes, err := e.AllocatedVolumes(); err != nil {
		log.Error.Printf("Unable to find allocated volumes : %s", err)
	} else {

		done := make(chan int)
		defer close(done)

		for _, volume := range volumes {

			go func(volume *AllocatedVolume) {

				if err := volume.Attach(); err != nil {
					log.Error.Printf("Unable to attach volume : %s\n", err)
				}
				done <- 1
			}(volume)

		}

		for i := 0; i < len(volumes); i++ {
			<-done
		}
	}

}
