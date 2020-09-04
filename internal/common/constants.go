package common

import "time"

const (
	CmpAPIGroup        = "compliance.openshift.io"
	CmpResourceVersion = "v1alpha1"
	RetryInterval      = time.Second * 2
	Timeout            = time.Minute * 20
)
