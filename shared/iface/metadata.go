package iface

type Metadata interface {
	InstanceID() (string, error)
	Region() (string, error)
}
