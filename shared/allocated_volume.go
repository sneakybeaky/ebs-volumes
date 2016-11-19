package shared

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/private/waiter"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared/log"
	"io"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type AllocatedVolume struct {
	VolumeId   string
	DeviceName string
	InstanceId string
	svc        ec2iface.EC2API
}

func NewAllocatedVolume(volumeId string, deviceName string, instanceId string, svc ec2iface.EC2API) *AllocatedVolume {

	return &AllocatedVolume{VolumeId: volumeId, DeviceName: deviceName, InstanceId: instanceId, svc: svc}
}

func (volume AllocatedVolume) String() string {
	return fmt.Sprintf("AllocatedVolume{ VolumeId : %s, DeviceName : %s, InstanceId : %s}", volume.VolumeId, volume.DeviceName, volume.InstanceId)
}

func (volume AllocatedVolume) Attach() error {

	log.Info.Printf("Attaching Volume (%s) at (%s)\n", volume.VolumeId, volume.DeviceName)

	if err := volume.waitUntilAvailable(); err != nil {
		return fmt.Errorf("Error waiting for Volume (%s) to become available, error: %s",
			volume.VolumeId, err)
	}

	opts := &ec2.AttachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceId),
		VolumeId:   aws.String(volume.VolumeId),
	}

	if _, err := volume.svc.AttachVolume(opts); err != nil {

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

func (volume AllocatedVolume) Detach() error {

	log.Info.Printf("Detaching Volume (%s) from (%s)\n", volume.VolumeId, volume.DeviceName)

	opts := &ec2.DetachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceId),
		VolumeId:   aws.String(volume.VolumeId),
	}

	if _, err := volume.svc.DetachVolume(opts); err != nil {

		if awsErr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("Error Detaching volume (%s) to instance (%s), message: \"%s\", code: \"%s\"",
				volume.VolumeId, volume.InstanceId, awsErr.Message(), awsErr.Code())
		}

	} else {

		err := volume.waitUntilDetached()

		if err != nil {
			return fmt.Errorf("Error waiting for Volume (%s) to detach at (%s), error: %s",
				volume.VolumeId, volume.DeviceName, err)
		} else {
			log.Info.Printf("Detached Volume (%s) from (%s)\n", volume.VolumeId, volume.DeviceName)
		}
	}

	return nil

}

func (volume AllocatedVolume) Info(w io.Writer) error {

	if status, err := volume.svc.DescribeVolumes(volume.describeVolumesInputWhenDetached()); err != nil {

		if awsErr, ok := err.(awserr.Error); ok {
			return fmt.Errorf("Error getting volume status for Volume (%s), message: \"%s\", code: \"%s\"",
				volume.VolumeId, awsErr.Message(), awsErr.Code())
		}

	} else {
		volumeStatus := status.Volumes[0]
		fmt.Fprintf(w, "Volume ID (%s), Device Name (%s), Status is %s\n",
			volume.VolumeId, volume.DeviceName, *volumeStatus.State)
	}

	return nil
}

func (volume AllocatedVolume) describeVolumesInputWhenAttached() *ec2.DescribeVolumesInput {
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

func (volume AllocatedVolume) describeVolumesInputWhenDetached() *ec2.DescribeVolumesInput {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volume.VolumeId)},
	}
}

func (volume AllocatedVolume) waitUntilAvailable() error {

	log.Debug.Printf("Waiting for volume (%s) to become available\n", volume.VolumeId)
	return volume.svc.WaitUntilVolumeAvailable(volume.describeVolumesInputWhenDetached())
}

// waitUntilVolumeAttached uses the Amazon EC2 API operation
// DescribeVolumes to wait for a condition to be met before returning.
// If the condition is not meet within the max attempt window an error will
// be returned.
func (volume AllocatedVolume) waitUntilAttached() error {

	input := volume.describeVolumesInputWhenAttached()

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
		Client: volume.svc,
		Input:  input,
		Config: waiterCfg,
	}

	log.Debug.Printf("Waiting for volume (%s) to be attached at (%s)\n", volume.VolumeId, volume.DeviceName)

	return w.Wait()
}

// waitUntilVolumeDetached uses the Amazon EC2 API operation
// DescribeVolumes to wait for a condition to be met before returning.
// If the condition is not meet within the max attempt window an error will
// be returned.
func (volume AllocatedVolume) waitUntilDetached() error {

	input := volume.describeVolumesInputWhenDetached()

	waiterCfg := waiter.Config{
		Operation:   "DescribeVolumes",
		Delay:       15,
		MaxAttempts: 40,
		Acceptors: []waiter.WaitAcceptor{
			{
				State:    "success",
				Matcher:  "pathAll",
				Argument: "Volumes[].State",
				Expected: ec2.VolumeStateAvailable,
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Volumes[].Attachments[].State",
				Expected: ec2.AttachmentStatusAttached,
			},
		},
	}

	w := waiter.Waiter{
		Client: volume.svc,
		Input:  input,
		Config: waiterCfg,
	}

	log.Debug.Printf("Waiting for volume (%s) to be detached from (%s)\n", volume.VolumeId, volume.DeviceName)

	return w.Wait()
}
