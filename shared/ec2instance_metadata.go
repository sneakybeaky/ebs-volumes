package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// A EC2InstanceMetadata provides metadata about an EC2 instance.
type EC2InstanceMetadata struct {
	EC2Metadata *ec2metadata.EC2Metadata
}

// NewEC2InstanceMetadata returns a new instance
func NewEC2InstanceMetadata(session *session.Session, cfgs ...*aws.Config) *EC2InstanceMetadata {
	return &EC2InstanceMetadata{
		EC2Metadata: ec2metadata.New(session, cfgs...),
	}

}

// InstanceID returns the instance id for this EC2 instance
func (e EC2InstanceMetadata) InstanceID() (string, error) {
	doc, err := e.EC2Metadata.GetInstanceIdentityDocument()

	if err != nil {
		return "", err
	}

	return doc.InstanceID, nil
}

// Region returns the region id for this EC2 instance
func (e EC2InstanceMetadata) Region() (string, error) {
	region, err := e.EC2Metadata.Region()

	if err != nil {
		return "", err
	}

	return region, nil
}
