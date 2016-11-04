package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sneakybeaky/aws-volumes/shared"
	"log"
)

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create session %v\n", err)
	}

	ec2Identity := shared.NewEC2Identity(sess)

	id, err := ec2Identity.GetInstanceID()

	if err != nil {
		fmt.Printf("My id is %s \n", id)
	}
}
