package util

const (
	OperatorNamespace         = "openshift-cluster-csi-drivers"
	CloudCredentialSecretName = "cloud-credentials"
	ManilaSecretName          = "manila-credentials"

	// Coordinates of config map with OpenStack certificate
	CloudConfigNamespace = "openshift-config"
	CloudConfigName      = "cloud-provider-config"

	StorageClassNamePrefix = "csi-manila-"

	// OpenStack config file name (as present in the operator Deployment)
	CloudConfigFilename = "/etc/clouds.yaml"

	// Name of cloud in secret provided by cloud-credentials-operator
	CloudName = "openstack"
)
