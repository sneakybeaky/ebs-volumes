package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sneakybeaky/aws-volumes/shared"
	"log"
)

func findAttachedVolumesFor(session *session.Session, instanceid string) ([]*ec2.Volume, error) {

	svc := ec2.New(session)

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

	resp, err := svc.DescribeVolumes(params)

	if err != nil {
		return nil, err
	}

	return resp.Volumes, nil

}

func detachVolumes(volumes []*ec2.Volume) {
	fmt.Println("Following volumes are attached")
	for volume := range volumes {
		fmt.Println(volume)
	}
}

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create session %v\n", err)
	}

	ec2Instance := shared.NewEC2Instance(sess)

	id, err := ec2Instance.InstanceID()

	if err != nil {
		log.Fatalf("failed to get instance id %v\n", err)
	}
	fmt.Printf("%s\n", id)

	region,err := ec2Instance.Region()
	if err != nil {
		log.Fatalf("failed to get region %v\n", err)
	}

	sess.Config.Region = &region
	volumes, err := findAttachedVolumesFor(sess, id)

	if err != nil {
		log.Fatalf("failed to find attached volumes %v\n", err)
	}

	detachVolumes(volumes)

}
