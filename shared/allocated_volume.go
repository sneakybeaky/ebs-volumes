package shared

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/aws-volumes/shared/log"
)

type AllocatedVolume struct {
	VolumeId   string
	DeviceName string
	InstanceId string
	svc        ec2iface.EC2API
}

// NewAllocatedVolume returns a new instance of AllocatedVolume
func NewAllocatedVolume(volumeId string, deviceName string, instanceId string, svc ec2iface.EC2API) *AllocatedVolume {

	return &AllocatedVolume{VolumeId: volumeId, DeviceName: deviceName, InstanceId: instanceId, svc: svc}
}

func (volume AllocatedVolume) String() string {
	return fmt.Sprintf("AllocatedVolume{ VolumeId : %s, DeviceName : %s, InstanceId : %s}", volume.VolumeId, volume.DeviceName, volume.InstanceId)
}

// Attach attempts to attach the volume
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

		return fmt.Errorf("Error attaching volume (%s) to instance (%s), cause: \"%s\"",
			volume.VolumeId, volume.InstanceId, err.Error())

	} else {

		err := volume.waitUntilAttached()

		if err != nil {
			return fmt.Errorf("Error waiting for Volume (%s) to attach at (%s), error: %s",
				volume.VolumeId, volume.DeviceName, err)
		}

		log.Info.Printf("Attached Volume (%s) at (%s)\n", volume.VolumeId, volume.DeviceName)

	}

	return nil

}

// Detach attempts to detach the volume
func (volume AllocatedVolume) Detach() error {

	log.Info.Printf("Detaching Volume (%s) from (%s)\n", volume.VolumeId, volume.DeviceName)

	opts := &ec2.DetachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceId),
		VolumeId:   aws.String(volume.VolumeId),
	}

	if _, err := volume.svc.DetachVolume(opts); err != nil {

		return fmt.Errorf("Error Detaching volume (%s) to instance (%s), cause : \"%s\"",
			volume.VolumeId, volume.InstanceId, err.Error())

	} else {

		err := volume.waitUntilAvailable()

		if err != nil {
			return fmt.Errorf("Error waiting for Volume (%s) to detach at (%s), cause: %s",
				volume.VolumeId, volume.DeviceName, err.Error())
		} else {
			log.Info.Printf("Detached Volume (%s) from (%s)\n", volume.VolumeId, volume.DeviceName)
		}
	}

	return nil

}

func (volume AllocatedVolume) Info(w io.Writer) error {

	if status, err := volume.svc.DescribeVolumes(volume.describeVolumesInputWhenDetached()); err != nil {

		return fmt.Errorf("Error getting volume status for Volume (%s), cause: \"%s\"",
			volume.VolumeId, err.Error())

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
			&ec2.Filter{
				Name:   aws.String("attachment.status"),
				Values: []*string{aws.String(ec2.AttachmentStatusAttached)},
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

	log.Debug.Printf("Waiting for volume (%s) to be attached at (%s)\n", volume.VolumeId, volume.DeviceName)

	return volume.svc.WaitUntilVolumeInUse(input)

}
