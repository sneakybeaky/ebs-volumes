package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
	"fmt"
	"log"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
)

type AllocatedVolume struct {
	VolumeID   string
	DeviceName string
	InstanceId string
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
				allocated = append(allocated, AllocatedVolume{VolumeID: *tag.Value, DeviceName: device, InstanceId : *tag.ResourceId})
			}
		}
	}

	return allocated, nil
}

func (e EC2Instance) AttachAllocatedVolumes() error {

	if allocated, err := e.AllocatedVolumes(); err != nil {
		return fmt.Errorf("[WARN] Error finding allocated volumes : %#v", err)
	} else {

		for _, allocated := range allocated {

			opts := &ec2.AttachVolumeInput{
				Device:     aws.String(allocated.DeviceName),
				InstanceId: aws.String(allocated.InstanceId),
				VolumeId:   aws.String(allocated.VolumeID),
			}

			log.Printf("[DEBUG] Attaching Volume (%s) to Instance (%s)", allocated.VolumeID, allocated.InstanceId)

			if _, err := e.EC2.AttachVolume(opts); err != nil {

				if awsErr, ok := err.(awserr.Error); ok {
					log.Printf("[WARN] Error attaching volume (%s) to instance (%s), message: \"%s\", code: \"%s\"",
						allocated.VolumeID, allocated.InstanceId, awsErr.Message(), awsErr.Code())
				}

			} else {

				stateConf := &resource.StateChangeConf{
					Pending:    []string{"attaching"},
					Target:     []string{"attached"},
					Refresh:    volumeAttachmentStateRefreshFunc(e.EC2, allocated.VolumeID, allocated.InstanceId),
					Timeout:    5 * time.Minute,
					Delay:      10 * time.Second,
					MinTimeout: 3 * time.Second,
				}

				_, err = stateConf.WaitForState()
				if err != nil {
					log.Printf(
						"Error waiting for Volume (%s) to attach to Instance: %s, error: %s",
						allocated.VolumeID, allocated.InstanceId, err)
				}
			}

		}
	}

	return nil

}

func volumeAttachmentStateRefreshFunc(conn *ec2.EC2, volumeID, instanceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(volumeID)},
			Filters: []*ec2.Filter{
				&ec2.Filter{
					Name:   aws.String("attachment.instance-id"),
					Values: []*string{aws.String(instanceID)},
				},
			},
		}

		resp, err := conn.DescribeVolumes(request)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				return nil, "failed", fmt.Errorf("code: %s, message: %s", awsErr.Code(), awsErr.Message())
			}
			return nil, "failed", err
		}

		if len(resp.Volumes) > 0 {
			v := resp.Volumes[0]
			for _, a := range v.Attachments {
				if a.InstanceId != nil && *a.InstanceId == instanceID {
					return a, *a.State, nil
				}
			}
		}
		// assume detached if volume count is 0
		return 42, "detached", nil
	}
}

