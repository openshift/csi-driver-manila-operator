package util

const (
	OperatorNamespace         = "openshift-cluster-csi-drivers"
	CloudCredentialSecretName = "manila-cloud-credentials"
	ManilaSecretName          = "manila-driver-credentials"

	StorageClassNamePrefix = "csi-manila-"

	// OpenStack config file name (as present in the operator Deployment)
	CloudConfigFilename = "/etc/openstack/clouds.yaml"

	// Name of cloud in secret provided by cloud-credentials-operator
	CloudName = "openstack"
)
