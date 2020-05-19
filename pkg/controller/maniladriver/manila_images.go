package maniladriver

import (
	"os"
)

const (
	defaultExternalProvisionerImage    = "quay.io/openshift/origin-csi-external-provisioner:latest"
	defaultExternalSnaphotterImage     = "quay.io/openshift/origin-csi-external-snapshotter:latest"
	defaultCSIDriverManilaImage        = "quay.io/openshift/origin-csi-driver-manila:latest"
	defaultCSINodeDriverRegistrarImage = "quay.io/openshift/origin-csi-node-driver-registrar:latest"
	defaultCSIDriverNFSImage           = "quay.io/openshift/origin-csi-driver-nfs:latest"

	externalProvisionerImageEnv    = "EXTERNAL_PROVISIONER_IMAGE"
	externalSnaphotterImageEnv     = "EXTERNAL_SNAPSHOTTER_IMAGE"
	csiDriverManilaImageEnv        = "CSI_DRIVER_MANILA_IMAGE"
	csiNodeDriverRegistrarImageEnv = "CSI_NODE_DRIVER_REGISTRAR_IMAGE"
	csiDriverNFSImage              = "CSI_DRIVER_NFS_IMAGE"
)

func getExternalProvisionerImage() string {
	if externalProvisionerImageFromEnv := os.Getenv(externalProvisionerImageEnv); externalProvisionerImageFromEnv != "" {
		return externalProvisionerImageFromEnv
	}
	return defaultExternalProvisionerImage
}

func getExternalSnaphotterImage() string {
	if externalSnaphotterImageFromEnv := os.Getenv(externalSnaphotterImageEnv); externalSnaphotterImageFromEnv != "" {
		return externalSnaphotterImageFromEnv
	}
	return defaultExternalSnaphotterImage
}

func getCSIDriverManilaImage() string {
	if csiDriverManilaImageFromEnv := os.Getenv(csiDriverManilaImageEnv); csiDriverManilaImageFromEnv != "" {
		return csiDriverManilaImageFromEnv
	}
	return defaultCSIDriverManilaImage
}

func getCSINodeDriverRegistrarImage() string {
	if csiNodeDriverRegistrarImageFromEnv := os.Getenv(csiNodeDriverRegistrarImageEnv); csiNodeDriverRegistrarImageFromEnv != "" {
		return csiNodeDriverRegistrarImageFromEnv
	}
	return defaultCSINodeDriverRegistrarImage
}

func getCSIDriverNFSImage() string {
	if csiDriverNFSImageFromEnv := os.Getenv(csiDriverNFSImage); csiDriverNFSImageFromEnv != "" {
		return csiDriverNFSImageFromEnv
	}
	return defaultCSIDriverNFSImage
}
