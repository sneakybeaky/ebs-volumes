package iface

// Metadata returns information about an EC2 instance
type Metadata interface {
	InstanceID() (string, error)
	Region() (string, error)
}
