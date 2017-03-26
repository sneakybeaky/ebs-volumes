package shared

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/aws-volumes/shared/iface"
	"github.com/sneakybeaky/aws-volumes/shared/log"
)

// VolumeTagPrefix prefixes the name of a tag describing an allocated volume
const VolumeTagPrefix = "volume_"

// DetachVolumesTag when set to a true value signals volumes can be detached
const DetachVolumesTag = "detach_volumes"

// EC2Instance provides metadata about an EC2 instance.
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

	tags, err := e.Tags()

	if err != nil {
		return allocated, err
	}

	for _, tag := range tags {
		if strings.HasPrefix(*tag.Key, VolumeTagPrefix) {

			key := *tag.Key
			device := key[len(VolumeTagPrefix):]
			allocated = append(allocated, NewAllocatedVolume(*tag.Value, device, *tag.ResourceId, e.svc))
		}
	}

	return allocated, nil
}

func (e EC2Instance) DetachVolumes() error {

	tags, err := e.Tags()

	if err != nil {
		return err
	}

	for _, tag := range tags {
		if *tag.Key == DetachVolumesTag {

			detachVolumes, _ := strconv.ParseBool(*tag.Value)

			if detachVolumes {
				e.applyToVolumes(detachVolume)
			} else {
				log.Info.Printf("Tag '%s' value is '%s' - not detaching volumes", DetachVolumesTag, *tag.Value)
			}

			break

		}
	}

	return nil
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

	if err := volume.Info(buf); err != nil {
		log.Error.Printf("Unable to get info for volume : %s\n", err)
		return
	}
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
