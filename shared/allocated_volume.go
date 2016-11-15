package shared

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"log"
	"time"
	"io"
)

type AllocatedVolume struct {
	VolumeId   string
	DeviceName string
	InstanceId string
	EC2        *ec2.EC2
}

func NewAllocatedVolume(volumeId string, deviceName string, instanceId string, EC2 *ec2.EC2) *AllocatedVolume {

	return &AllocatedVolume{VolumeId: volumeId, DeviceName: deviceName, InstanceId: instanceId, EC2: EC2}
}

func (volume AllocatedVolume) Attach() error {

	opts := &ec2.AttachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceId),
		VolumeId:   aws.String(volume.VolumeId),
	}

	log.Printf("[DEBUG] Attaching Volume (%s) to Instance (%s)", volume.VolumeId, volume.InstanceId)

	if _, err := volume.EC2.AttachVolume(opts); err != nil {

		if awsErr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("Error attaching volume (%s) to instance (%s), message: \"%s\", code: \"%s\"",
				volume.VolumeId, volume.InstanceId, awsErr.Message(), awsErr.Code())
		}

	} else {

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"attaching"},
			Target:     []string{"attached"},
			Refresh:    volumeAttachmentStateRefreshFunc(volume.EC2, volume.VolumeId, volume.InstanceId),
			Timeout:    5 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("Error waiting for Volume (%s) to attach to Instance: %s, error: %s",
				volume.VolumeId, volume.InstanceId, err)
		}
	}

	return nil

}

func (volume AllocatedVolume) Info(w io.Writer) {
	fmt.Fprintf(w,"Instance ID %s, Device Name %s, Volume ID %s\n",volume.InstanceId,volume.DeviceName, volume.VolumeId)
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
