package shared

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/aws-volumes/shared/iface"
	"github.com/sneakybeaky/aws-volumes/shared/log"
	"os"
	"strings"
	"sync"
)

const volume_tag_prefix = "volume_"

// A EC2InstanceMetadata provides metadata about an EC2 instance.
type EC2Instance struct {
	svc      ec2iface.EC2API
	metadata iface.Metadata
}

func NewEC2Instance(metadata iface.Metadata, svc ec2iface.EC2API) *EC2Instance {

	return &EC2Instance{
		svc:      svc,
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
	resp, err := e.svc.DescribeTags(params)

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
				allocated = append(allocated, NewAllocatedVolume(*tag.Value, device, *tag.ResourceId, e.svc))
			}
		}
	}

	return allocated, nil
}

func (e EC2Instance) DetachVolumes() {
	e.applyToVolumes(detachVolume)
}

func (e EC2Instance) AttachVolumes() {
	e.applyToVolumes(attachVolume)
}

func (e EC2Instance) ShowVolumesInfo() {
	e.applyToVolumes(showVolumeInfo)
}

var attachVolume = func(volume *AllocatedVolume) {

	if err := volume.Attach(); err != nil {
		log.Error.Printf("Unable to attach volume : %s\n", err)
	}
}

var detachVolume = func(volume *AllocatedVolume) {

	if err := volume.Detach(); err != nil {
		log.Error.Printf("Unable to detach volume : %s\n", err)
	}
}

var showVolumeInfo = func(volume *AllocatedVolume) {
	buf := new(bytes.Buffer)
	volume.Info(buf)
	os.Stdout.WriteString(buf.String())
}

func (e EC2Instance) applyToVolumes(action func(volume *AllocatedVolume)) {
	if volumes, err := e.AllocatedVolumes(); err != nil {
		log.Error.Printf("Unable to find allocated volumes : %s", err)
	} else {

		var wg sync.WaitGroup

		for _, volume := range volumes {

			wg.Add(1)
			go func(action func(volume *AllocatedVolume), volume *AllocatedVolume) {

				defer wg.Done()
				action(volume)

			}(action, volume)

		}

		wg.Wait()
	}

}
