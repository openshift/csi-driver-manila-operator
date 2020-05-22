package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaDriver) handleManilaDriverNamespace(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Driver Namespace")

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "openshift-manila-csi-driver",
		},
	}

	// Check if this Namespace already exists
	found := &corev1.Namespace{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: ns.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", "Namespace.Name", ns.Name)
		err = r.client.Create(context.TODO(), ns)
		if err != nil {
			return err
		}

		// Namespace created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Namespace already exists - don't requeue
	reqLogger.Info("Skip reconcile: Namespace already exists", "Namespace.Name", found.Name)
	return nil
}
