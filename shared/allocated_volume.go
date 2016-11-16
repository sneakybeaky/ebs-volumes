package shared

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"log"
	"io"
	"github.com/aws/aws-sdk-go/private/waiter"
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

		err := volume.waitUntilVolumeAttached()

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

// waitUntilVolumeAttached uses the Amazon EC2 API operation
// DescribeVolumes to wait for a condition to be met before returning.
// If the condition is not meet within the max attempt window an error will
// be returned.
func (volume AllocatedVolume) waitUntilVolumeAttached() error {

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
				Expected: "attached",
			},
			{
				State:    "failure",
				Matcher:  "pathAny",
				Argument: "Volumes[].Attachments[].State",
				Expected: "detached",
			},
		},
	}

	w := waiter.Waiter{
		Client: volume.EC2,
		Input:  input,
		Config: waiterCfg,
	}
	return w.Wait()
}