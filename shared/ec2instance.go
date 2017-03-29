package shared

import (
	"bytes"
	"os"
	"strconv"
	"strings"
	"sync"

	"fmt"

	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/ebs-volumes/shared/iface"
	"github.com/sneakybeaky/ebs-volumes/shared/log"
)

// VolumeTagPrefix prefixes the name of a tag describing an allocated volume
const VolumeTagPrefix = "volume_"

// DetachVolumesTag when set to a true value signals volumes can be detached
const DetachVolumesTag = "detach_volumes"

// GetInstance returns a representation for the current EC2 instance
func GetInstance() (*EC2Instance, error) {

	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed to create AWS session : %v", err)
	}

	metadata := NewEC2InstanceMetadata(sess)

	region, err := metadata.Region()

	if err != nil {
		return nil, fmt.Errorf("Failed to get AWS region : %v", err)
	}

	sess.Config.Region = &region

	return NewEC2Instance(metadata, ec2.New(sess)), nil

}

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
				err := e.applyToVolumes(detachVolume)
				if err != nil {
					return err
				}
			} else {
				log.Info.Printf("Tag '%s' value is '%s' - not detaching volumes", DetachVolumesTag, *tag.Value)
			}

			break

		}
	}

	return nil
}

func (e EC2Instance) AttachVolumes() error {
	return e.applyToVolumes(attachVolume)
}

func (e EC2Instance) ShowVolumesInfo() {
	e.applyToVolumes(showVolumeInfo)
}

var attachVolume = func(volume *AllocatedVolume) error {

	if err := volume.Attach(); err != nil {
		log.Error.Printf("Unable to attach volume : %s\n", err)
	}

	return nil
}

var detachVolume = func(volume *AllocatedVolume) error {

	if err := volume.Detach(); err != nil {
		return fmt.Errorf("Unable to detach volume : %v\n", err)
	}
	return nil
}

var showVolumeInfo = func(volume *AllocatedVolume) error {
	buf := new(bytes.Buffer)

	if err := volume.Info(buf); err != nil {
		log.Error.Printf("Unable to get info for volume : %s\n", err)
		return nil
	}
	os.Stdout.WriteString(buf.String())

	return nil
}

func (e EC2Instance) applyToVolumes(action func(volume *AllocatedVolume) error) error {
	if volumes, err := e.AllocatedVolumes(); err != nil {
		return fmt.Errorf("Unable to find allocated volumes : %v", err)
	} else {

		var wg sync.WaitGroup

		failed := false

		for _, volume := range volumes {

			wg.Add(1)
			go func(action func(volume *AllocatedVolume) error, volume *AllocatedVolume) {

				defer wg.Done()
				err := action(volume)

				if err != nil {
					log.Error.Println(err)
					failed = true
				}

			}(action, volume)

		}

		wg.Wait()

		if failed {
			return errors.New("Failed for some volumes")
		}
	}

	return nil
}
