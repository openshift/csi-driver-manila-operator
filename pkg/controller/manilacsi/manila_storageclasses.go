package manilacsi

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	storageClassNamePrefix = "csi-manila-"
)

func (r *ReconcileManilaCSI) handleManilaStorageClasses(instance *manilacsiv1alpha1.ManilaCSI, shareTypes []sharetypes.ShareType, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila StorageClasses")

	for _, shareType := range shareTypes {
		err := r.handleManilaStorageClass(instance, shareType, reqLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileManilaCSI) handleManilaStorageClass(instance *manilacsiv1alpha1.ManilaCSI, shareType sharetypes.ShareType, reqLogger logr.Logger) error {
	storageClassName := storageClassNamePrefix + shareType.Name
	reqLogger.Info("Reconciling Manila StorageClass", "StorageClass.Name", storageClassName)

	// Define a new StorageClass object
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: storageClassName,
		},
		Provisioner: "manila.csi.openstack.org",
		Parameters: map[string]string{
			"type": shareType.Name,
			"csi.storage.k8s.io/provisioner-secret-name":       "csi-manila-secrets",
			"csi.storage.k8s.io/provisioner-secret-namespace":  "manila-csi",
			"csi.storage.k8s.io/node-stage-secret-name":        "csi-manila-secrets",
			"csi.storage.k8s.io/node-stage-secret-namespace":   "manila-csi",
			"csi.storage.k8s.io/node-publish-secret-name":      "csi-manila-secrets",
			"csi.storage.k8s.io/node-publish-secret-namespace": "manila-csi",
		},
	}

	// Check if this StorageClass already exists
	found := &storagev1.StorageClass{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: sc.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new StorageClass", "StorageClass.Name", sc.Name)
		err = r.client.Create(context.TODO(), sc)
		if err != nil {
			return err
		}

		// StorageClass created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// StorageClass already exists - don't requeue
	reqLogger.Info("Skip reconcile: StorageClass already exists", "StorageClass.Name", found.Name)
	return nil
}
