package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sneakybeaky/aws-volumes/shared"
	"log"
)

func listVolumes(instance *shared.EC2Instance) {

	volumes, err := instance.AttachedVolumes()

	if err != nil {
		log.Fatalf("failed to find attached volumes %v\n", err)
	}

	fmt.Println("Following volumes are attached")
	for _, volume := range volumes {
		fmt.Println(volume)
	}
}

func listTags(instance *shared.EC2Instance) {
	tags, err := instance.Tags()

	if err != nil {
		log.Fatalf("failed to find tags %v\n", err)
	}

	fmt.Println("Following tags found")
	for _, tag := range tags {
		fmt.Println(tag)
	}
}

func listAllocatedVolumes(instance *shared.EC2Instance) {
	volumes, err := instance.AllocatedVolumes()

	if err != nil {
		log.Fatalf("failed to find allocated volumes %v\n", err)
	}

	fmt.Println("Following volumes found")
	for _, volume := range volumes {
		fmt.Printf("%+v\n", *volume)
	}
}



func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create session %v\n", err)
	}

	metadata := shared.NewEC2InstanceMetadata(sess)

	id, err := metadata.InstanceID()

	if err != nil {
		log.Fatalf("failed to get instance id %v\n", err)
	}
	fmt.Printf("%s\n", id)

	region, err := metadata.Region()
	if err != nil {
		log.Fatalf("failed to get region %v\n", err)
	}

	sess.Config.Region = &region

	instance := shared.NewEC2Instance(metadata, sess)

	listVolumes(instance)

	listTags(instance)

	listAllocatedVolumes(instance)

}
