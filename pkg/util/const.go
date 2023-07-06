package util

const (
	OperatorNamespace         = "openshift-cluster-csi-drivers"
	OperandNamespace          = "openshift-manila-csi-driver"
	CloudCredentialSecretName = "manila-cloud-credentials"
	ManilaSecretName          = "csi-manila-secrets"

	CloudConfigNamespace = "openshift-config"
	CloudConfigName      = "cloud-provider-config"

	StorageClassNamePrefix = "csi-manila-"

	// Name of cloud in secret provided by cloud-credentials-operator
	CloudName = "openstack"
)
