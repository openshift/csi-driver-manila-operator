package maniladriver

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	storageClassNamePrefix = "csi-manila-"
)

func (r *ReconcileManilaDriver) handleManilaStorageClasses(instance *maniladriverv1alpha1.ManilaDriver, shareTypes []sharetypes.ShareType, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila StorageClasses")

	for _, shareType := range shareTypes {
		err := r.handleManilaStorageClass(instance, shareType, reqLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileManilaDriver) handleManilaStorageClass(instance *maniladriverv1alpha1.ManilaDriver, shareType sharetypes.ShareType, reqLogger logr.Logger) error {
	storageClassName := storageClassNamePrefix + strings.ToLower(shareType.Name)
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
			"csi.storage.k8s.io/provisioner-secret-namespace":  "openshift-manila-csi-driver",
			"csi.storage.k8s.io/node-stage-secret-name":        "csi-manila-secrets",
			"csi.storage.k8s.io/node-stage-secret-namespace":   "openshift-manila-csi-driver",
			"csi.storage.k8s.io/node-publish-secret-name":      "csi-manila-secrets",
			"csi.storage.k8s.io/node-publish-secret-namespace": "openshift-manila-csi-driver",
		},
	}

	if err := annotator.SetLastAppliedAnnotation(sc); err != nil {
		return err
	}

	// Check if this StorageClass already exists
	found := &storagev1.StorageClass{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: sc.Name, Namespace: ""}, found)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err == nil {
		// Check if we need to update the object
		equal, err := compareLastAppliedAnnotations(found, sc)
		if err != nil {
			return err
		}

		if !equal {
			// StorageClass can't be updated directly, so we have to delete it and create again
			reqLogger.Info("Deleting StorageClass", "StorageClass.Name", found.Name)
			err = r.client.Delete(context.TODO(), found)
			if err != nil {
				return err
			}
		} else {
			// StorageClass already exists - don't requeue
			reqLogger.Info("Skip reconcile: StorageClass already exists", "StorageClass.Name", found.Name)
			return nil
		}
	}

	reqLogger.Info("Creating a new StorageClass", "StorageClass.Name", sc.Name)
	err = r.client.Create(context.TODO(), sc)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManilaDriver) deleteManilaStorageClasses(reqLogger logr.Logger) error {
	reqLogger.Info("Deleting Manila StorageClasses")

	scs := &storagev1.StorageClassList{}
	err := r.apiReader.List(context.TODO(), scs, &client.ListOptions{})
	if err != nil {
		return err
	}

	for _, sc := range scs.Items {
		if sc.Provisioner == "manila.csi.openstack.org" {
			err = r.client.Delete(context.TODO(), &sc)
			if err != nil {
				return err
			}

			reqLogger.Info("Storage Class was deleted succesfully", "StorageClass.Name", sc.Name)
		}
	}

	return nil
}
