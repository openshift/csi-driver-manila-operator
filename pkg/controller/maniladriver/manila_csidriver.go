package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// falsePTR returns a *bool whose underlying value is false.
func falsePTR() *bool {
	t := false
	return &t
}

func (r *ReconcileManilaDriver) handleManilaCSIDriver(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila CSIDriver")

	// Define a new CSIDriver object
	driver := generateCSIDriver()

	if err := annotator.SetLastAppliedAnnotation(driver); err != nil {
		return err
	}

	// Check if this CSIDriver already exists
	found := &storagev1beta1.CSIDriver{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: driver.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new CSIDriver", "CSIDriver.Name", driver.Name)
		err = r.client.Create(context.TODO(), driver)
		if err != nil {
			return err
		}

		// CSIDriver created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, driver)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating CSIDriver with new changes", "CSIDriver.Namespace", found.Namespace, "CSIDriver.Name", found.Name)
		err = r.client.Update(context.TODO(), driver)
		if err != nil {
			return err
		}
	} else {
		// CSIDriver already exists - don't requeue
		reqLogger.Info("Skip reconcile: CSIDriver already exists", "CSIDriver.Namespace", found.Namespace, "CSIDriver.Name", found.Name)
	}

	return nil
}

func (r *ReconcileManilaDriver) deleteCSIDriver(reqLogger logr.Logger) error {
	reqLogger.Info("Deleting CSI Driver")

	driver := generateCSIDriver()

	err := r.client.Delete(context.TODO(), driver)
	if err != nil {
		return err
	}

	reqLogger.Info("CSI Driver was deleted succesfully", "CSIDriver.Name", driver.Name)

	return nil
}

func generateCSIDriver() *storagev1beta1.CSIDriver {
	return &storagev1beta1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: "manila.csi.openstack.org",
		},
		Spec: storagev1beta1.CSIDriverSpec{
			AttachRequired: falsePTR(),
			PodInfoOnMount: falsePTR(),
		},
	}
}
