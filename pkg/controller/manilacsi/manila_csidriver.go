package manilacsi

import (
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
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

func (r *ReconcileManilaCSI) handleManilaCSIDriver(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila CSIDriver")

	// Define a new CSIDriver object
	driver := &storagev1beta1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name: "manila.csi.openstack.org",
		},
		Spec: storagev1beta1.CSIDriverSpec{
			AttachRequired: falsePTR(),
			PodInfoOnMount: falsePTR(),
		},
	}

	// Check if this CSIDriver already exists
	found := &storagev1beta1.CSIDriver{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: driver.Name, Namespace: ""}, found)
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

	// CSIDriver already exists - don't requeue
	reqLogger.Info("Skip reconcile: CSIDriver already exists", "CSIDriver.Namespace", found.Namespace, "CSIDriver.Name", found.Name)
	return nil
}
