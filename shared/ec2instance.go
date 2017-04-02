package shared

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

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
		return nil, fmt.Errorf("failed to create AWS session : %v", err)
	}

	metadata := NewEC2InstanceMetadata(sess)

	region, err := metadata.Region()

	if err != nil {
		return nil, fmt.Errorf("failed to get AWS region : %v", err)
	}

	sess.Config.Region = &region

	return NewEC2Instance(metadata, ec2.New(sess)), nil

}

// EC2Instance provides metadata about an EC2 instance.
type EC2Instance struct {
	svc      ec2iface.EC2API
	metadata iface.Metadata
}

// NewEC2Instance returns a new EC2Instance
func NewEC2Instance(metadata iface.Metadata, svc ec2iface.EC2API) *EC2Instance {

	return &EC2Instance{
		svc:      svc,
		metadata: metadata,
	}

}

func (e EC2Instance) tags() ([]*ec2.TagDescription, error) {

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

// AllocatedVolumes returns the volumes allocated to this instance
func (e EC2Instance) AllocatedVolumes() ([]*AllocatedVolume, error) {
	var allocated []*AllocatedVolume

	tags, err := e.tags()

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

//shouldDetachVolumes returns true if volumes should be detached, false otherwise
func (e EC2Instance) shouldDetachVolumes() (bool, error) {
	tags, err := e.tags()

	if err != nil {
		return false, err
	}

	shouldDetach := false

	for _, tag := range tags {
		if *tag.Key == DetachVolumesTag {

			shouldDetach, _ = strconv.ParseBool(*tag.Value)

			if !shouldDetach {

				log.Debug.Printf("Tag '%s' value is '%s' - not detaching volumes\n", DetachVolumesTag, *tag.Value)
			}

			break

		}
	}

	return shouldDetach, nil

}

// DetachVolumes attempts to detach the allocated volumes attached to this instance, if the necessary tag has been set
func (e EC2Instance) DetachVolumes() error {

	detachVolumes, err := e.shouldDetachVolumes()

	if err != nil {
		return err
	}

	if !detachVolumes {
		return nil
	}

	return e.applyToVolumes(detachVolume)

}

// AttachVolumes attempts to attach the allocated volumes
func (e EC2Instance) AttachVolumes() error {
	return e.applyToVolumes(attachVolume)
}

// ShowVolumesInfo prints information about the allocated volumes
func (e EC2Instance) ShowVolumesInfo() error {
	return e.applyToVolumes(showVolumeInfo)
}

var attachVolume = func(volume *AllocatedVolume) error {

	if err := volume.Attach(); err != nil {
		return fmt.Errorf("unable to attach volume : %v\n", err)
	}

	return nil
}

var detachVolume = func(volume *AllocatedVolume) error {

	if err := volume.Detach(); err != nil {
		return fmt.Errorf("unable to detach volume : %v\n", err)
	}
	return nil
}

var showVolumeInfo = func(volume *AllocatedVolume) error {
	buf := new(bytes.Buffer)

	if err := volume.Info(buf); err != nil {
		return fmt.Errorf("unable to get info for volume : %v\n", err)
	}
	os.Stdout.WriteString(buf.String())

	return nil
}

func (e EC2Instance) applyToVolumes(action func(volume *AllocatedVolume) error) error {

	volumes, err := e.AllocatedVolumes()

	if err != nil {
		return fmt.Errorf("unable to find allocated volumes : %v", err)
	}

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
		return errors.New("failed for some volumes")
	}

	return nil
}
