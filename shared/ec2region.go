package shared

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

// A EC2Region provides information about an EC2 region.
type EC2Region struct {
	EC2Metadata *ec2metadata.EC2Metadata
}

func NewEC2Region(session *session.Session, cfgs ...*aws.Config) *EC2Region {
	return &EC2Region{
		EC2Metadata: ec2metadata.New(session, cfgs...),
	}

}

// GetInstanceID returns the instance id for this EC2 instance
func (e EC2Region) Region() (string, error) {
	region, err := e.EC2Metadata.Region()

	if err != nil {
		return "", err
	}

	return region, nil
}
