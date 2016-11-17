package shared

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/private/waiter"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared/log"
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

	if err := volume.waitUntilAvailable(); err != nil {
		return fmt.Errorf("Error waiting for Volume (%s) to become available, error: %s",
			volume.VolumeId, err)
	}

	opts := &ec2.AttachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceId),
		VolumeId:   aws.String(volume.VolumeId),
	}

	log.Debug.Printf("Attaching Volume (%s) at (%s)\n", volume.VolumeId, volume.DeviceName)

	if _, err := volume.EC2.AttachVolume(opts); err != nil {

		if awsErr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("Error attaching volume (%s) to instance (%s), message: \"%s\", code: \"%s\"",
				volume.VolumeId, volume.InstanceId, awsErr.Message(), awsErr.Code())
		}

	} else {

		err := volume.waitUntilAttached()

		if err != nil {
			return fmt.Errorf("Error waiting for Volume (%s) to attach at (%s), error: %s",
				volume.VolumeId, volume.DeviceName, err)
		} else {
			log.Info.Printf("Attached Volume (%s) at (%s)\n", volume.VolumeId, volume.DeviceName)
		}
	}

	return nil

}

func (volume AllocatedVolume) Info(w io.Writer) {
	fmt.Fprintf(w, "Instance ID %s, Device Name %s, Volume ID %s\n", volume.InstanceId, volume.DeviceName, volume.VolumeId)
}

func (volume AllocatedVolume) describeVolumesInput() *ec2.DescribeVolumesInput {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volume.VolumeId)},
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(volume.InstanceId)},
			},
		},
	}
}

func (volume AllocatedVolume) waitUntilAvailable() error {

	log.Debug.Printf("Waiting for volume (%s) to become available\n", volume.InstanceId)
	return volume.EC2.WaitUntilVolumeAvailable(volume.describeVolumesInput())
}

// waitUntilVolumeAttached uses the Amazon EC2 API operation
// DescribeVolumes to wait for a condition to be met before returning.
// If the condition is not meet within the max attempt window an error will
// be returned.
func (volume AllocatedVolume) waitUntilAttached() error {

	input := volume.describeVolumesInput()

	waiterCfg := waiter.Config{
		Operation:   "DescribeVolumes",
		Delay:       15,
		MaxAttempts: 40,
		Acceptors: []waiter.WaitAcceptor{
			{
				State:    "success",
				Matcher:  "pathAll",
				Argument: "Volumes[].Attachments[].State",
				Expected: ec2.AttachmentStatusAttached,
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Volumes[].Attachments[].State",
				Expected: ec2.AttachmentStatusDetached,
			},
		},
	}

	w := waiter.Waiter{
		Client: volume.EC2,
		Input:  input,
		Config: waiterCfg,
	}

	log.Debug.Printf("Waiting for volume (%s) to be attached at (%s)\n", volume.InstanceId, volume.DeviceName)

	return w.Wait()
}
