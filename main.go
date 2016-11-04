package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
)

// A EC2Identity provides information about an EC2 instance.
type EC2Identity struct {
	EC2Metadata *ec2metadata.EC2Metadata
}

func newEC2Identity(session *session.Session, cfgs ...*aws.Config) *EC2Identity {
	return &EC2Identity{
		EC2Metadata: ec2metadata.New(session, cfgs...),
	}

}

// GetInstanceID returns the instance id for this EC2 instance
func (e EC2Identity) GetInstanceID() (string, error) {
	doc, err := e.EC2Metadata.GetInstanceIdentityDocument()

	if err != nil {
		return "", err
	}

	return doc.InstanceID, nil
}

func main() {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatalf("failed to create session %v\n", err)
	}

	ec2Identity := newEC2Identity(sess)

	id, err := ec2Identity.GetInstanceID()

	if err != nil {
		fmt.Printf("My id is %s \n", id)
	}
}
