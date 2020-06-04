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

func (r *ReconcileManilaDriver) getCloudProviderCert() (string, error) {
	cm := &corev1.ConfigMap{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: "cloud-provider-config", Namespace: "openshift-config"}, cm)
	if err != nil {
		return "", err
	}

	return string(cm.Data["ca-bundle.pem"]), nil
}

// handleCACertConfigMap converts the cloud provider configmap with the ca cert, if it is available,
// into the driver configmap
func (r *ReconcileManilaDriver) handleCACertConfigMap(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling CA Cert ConfigMap")

	cert, err := r.getCloudProviderCert()
	if err != nil {
		return err
	}

	// We don't have the certificate, so we don't need to create the configmap in for the driver
	if cert == "" {
		return nil
	}

	cm := generateCACertConfigMap(cert)

	if err := annotator.SetLastAppliedAnnotation(cm); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err = r.apiReader.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
	if err == nil {
		// Check if we need to update the object
		equal, err := compareLastAppliedAnnotations(found, cm)
		if err != nil {
			return err
		}

		if !equal {
			reqLogger.Info("Updating ConfigMap with new changes", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
			err = r.client.Update(context.TODO(), cm)
			if err != nil {
				return err
			}
		} else {
			// ConfigMap already exists - don't requeue
			reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", found.Namespace, "ConfigMap.Name", found.Name)
		}

		return nil
	}

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Convert the cloud provider configmap with the ca cert into driver configmap
	reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
	err = r.client.Create(context.TODO(), cm)
	if err != nil {
		return err
	}

	// ConfigMap created successfully - don't requeue
	return nil
}

func generateCACertConfigMap(cert string) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-certificates",
			Namespace: "openshift-manila-csi-driver",
		},
		Data: map[string]string{"cloud-provider-ca-bundle.pem": cert},
	}

	return &cm
}
