package shared

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sneakybeaky/ebs-volumes/shared/log"
)

// AllocatedVolume represents an EBS volume that's been allocated to an EC2 Instance
type AllocatedVolume struct {
	VolumeID   string
	DeviceName string
	InstanceID string
	svc        ec2iface.EC2API
}

// NewAllocatedVolume returns a new instance of AllocatedVolume
func NewAllocatedVolume(volumeID string, deviceName string, instanceID string, svc ec2iface.EC2API) *AllocatedVolume {

	return &AllocatedVolume{VolumeID: volumeID, DeviceName: deviceName, InstanceID: instanceID, svc: svc}
}

func (volume AllocatedVolume) String() string {
	return fmt.Sprintf("AllocatedVolume{ VolumeID : %s, DeviceName : %s, InstanceID : %s}", volume.VolumeID, volume.DeviceName, volume.InstanceID)
}

// Attach attempts to attach the volume
func (volume AllocatedVolume) Attach() error {

	log.Info.Printf("Attaching Volume (%s) at (%s)\n", volume.VolumeID, volume.DeviceName)

	attached, err := volume.Attached()
	if err != nil {
		return fmt.Errorf("error Attaching volume (%s) to instance (%s): %v",
			volume.VolumeID, volume.InstanceID, err)
	}

	if attached {
		log.Debug.Printf("Volume (%s) already attached - skipping\n", volume.VolumeID)
		return nil
	}

	if err := volume.waitUntilAvailable(); err != nil {
		return fmt.Errorf("error waiting for volume (%s) to become available: %v",
			volume.VolumeID, err)
	}

	opts := &ec2.AttachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceID),
		VolumeId:   aws.String(volume.VolumeID),
	}

	if _, err := volume.svc.AttachVolume(opts); err != nil {

		return fmt.Errorf("error attaching volume (%s) to instance (%s): %s",
			volume.VolumeID, volume.InstanceID, err)

	}

	err = volume.waitUntilAttached()

	if err != nil {
		return fmt.Errorf("error waiting for volume (%s) to attach at (%s): %v",
			volume.VolumeID, volume.DeviceName, err)
	}

	log.Info.Printf("Attached Volume (%s) at (%s)\n", volume.VolumeID, volume.DeviceName)

	return nil

}

// Detach attempts to detach the volume
func (volume AllocatedVolume) Detach() error {

	log.Info.Printf("Detaching Volume (%s) from (%s)\n", volume.VolumeID, volume.DeviceName)

	attached, err := volume.Attached()
	if err != nil {
		return fmt.Errorf("error Detaching volume (%s) from instance (%s): %v",
			volume.VolumeID, volume.InstanceID, err)
	}

	if !attached {
		log.Debug.Printf("Volume (%s) not attached - skipping\n", volume.VolumeID)
		return nil
	}

	opts := &ec2.DetachVolumeInput{
		Device:     aws.String(volume.DeviceName),
		InstanceId: aws.String(volume.InstanceID),
		VolumeId:   aws.String(volume.VolumeID),
	}

	if _, err := volume.svc.DetachVolume(opts); err != nil {

		return fmt.Errorf("error detaching volume (%s) from instance (%s): %s",
			volume.VolumeID, volume.InstanceID, err)

	}

	err = volume.waitUntilAvailable()

	if err != nil {
		return fmt.Errorf("error waiting for volume (%s) to detach at (%s): %v",
			volume.VolumeID, volume.DeviceName, err)
	}

	log.Info.Printf("Detached Volume (%s) from (%s)\n", volume.VolumeID, volume.DeviceName)

	return nil

}

// Attached returns true if the volume is attached to the designated instance, false otherwise.
func (volume AllocatedVolume) Attached() (bool, error) {
	return doAttached(&volume)
}

var doAttached = func(volume *AllocatedVolume) (bool, error) {
	status, err := volume.svc.DescribeVolumes(volume.describeVolumesInputWhenAttached())

	if err != nil {

		return false, fmt.Errorf("error getting volume status for volume (%s): %v",
			volume.VolumeID, err)

	}

	return len(status.Volumes) > 0, nil
}

// Info writes information about this volume
func (volume AllocatedVolume) Info(w io.Writer) error {

	status, err := volume.svc.DescribeVolumes(volume.describeVolumesInput())

	if err != nil {

		return fmt.Errorf("error getting volume status for volume (%s): %v",
			volume.VolumeID, err)

	}

	volumeStatus := status.Volumes[0]
	fmt.Fprintf(w, "Volume ID (%s), Device Name (%s), Status is %s\n",
		volume.VolumeID, volume.DeviceName, *volumeStatus.State)

	return nil
}

// describeVolumesInput provides the structure to describe this volume when attached to the designated EC2 instance
func (volume AllocatedVolume) describeVolumesInputWhenAttached() *ec2.DescribeVolumesInput {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volume.VolumeID)},
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(volume.InstanceID)},
			},
			{
				Name:   aws.String("attachment.status"),
				Values: []*string{aws.String(ec2.AttachmentStatusAttached)},
			},
		},
	}
}

// describeVolumesInput provides the structure to describe this volume
func (volume AllocatedVolume) describeVolumesInput() *ec2.DescribeVolumesInput {
	return &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(volume.VolumeID)},
	}
}

func (volume AllocatedVolume) waitUntilAvailable() error {

	log.Debug.Printf("Waiting for volume (%s) to become available\n", volume.VolumeID)
	return volume.svc.WaitUntilVolumeAvailable(volume.describeVolumesInput())
}

// waitUntilVolumeAttached uses the Amazon EC2 API operation
// DescribeVolumes to wait for a condition to be met before returning.
// If the condition is not meet within the max attempt window an error will
// be returned.
func (volume AllocatedVolume) waitUntilAttached() error {

	input := volume.describeVolumesInputWhenAttached()

	log.Debug.Printf("Waiting for volume (%s) to be attached at (%s)\n", volume.VolumeID, volume.DeviceName)

	return volume.svc.WaitUntilVolumeInUse(input)

}
