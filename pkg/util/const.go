package util

const (
	OperandNamespace          = "openshift-manila-csi-driver"
	CloudCredentialSecretName = "manila-cloud-credentials"
	ManilaSecretName          = "manila-driver-credentials"

	StorageClassNamePrefix = "csi-manila-"

	// OpenStack config file name (as present in the operator Deployment)
	CloudConfigFilename = "/etc/openstack/clouds.yaml"
	CertFile            = "/etc/openstack-ca/ca-bundle.pem"

	// Name of cloud in secret provided by cloud-credentials-operator
	CloudName = "openstack"
)
