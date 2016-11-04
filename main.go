package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
)

type EC2Identity struct {
	EC2Metadata *ec2metadata.EC2Metadata
}

func newEC2Identity(session *session.Session) *EC2Identity {
	return &EC2Identity{
		EC2Metadata: ec2metadata.New(session),
	}

}

func (self EC2Identity) GetInstanceID() (string, error) {
	doc, err := self.EC2Metadata.GetInstanceIdentityDocument()

	if (err != nil) {
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