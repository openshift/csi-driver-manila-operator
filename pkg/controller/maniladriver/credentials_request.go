package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	credsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaDriver) handleCredentialsRequest(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Credentials Request")

	// Define a new Credential Request object
	creq := generateCredentialsRequest()

	if err := annotator.SetLastAppliedAnnotation(creq); err != nil {
		return err
	}

	// Check if this Credential Request already exists
	found := &credsv1.CredentialsRequest{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: creq.Name, Namespace: creq.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new CredentialsRequest", "CredentialsRequest.Namespace", creq.Namespace, "CredentialsRequest.Name", creq.Name)
		err = r.client.Create(context.TODO(), creq)
		if err != nil {
			return err
		}

		// Credential Request created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, creq)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating CredentialsRequest with new changes", "CredentialsRequest.Namespace", found.Namespace, "CredentialsRequest.Name", found.Name)
		err = r.client.Update(context.TODO(), creq)
		if err != nil {
			return err
		}
	} else {
		// Credential Request already exists - don't requeue
		reqLogger.Info("Skip reconcile: CredentialsRequest already exists", "CredentialsRequest.Namespace", found.Namespace, "CredentialsRequest.Name", found.Name)
	}

	return nil
}

func generateCredentialsRequest() *credsv1.CredentialsRequest {
	openstackProvSpec := &credsv1.OpenStackProviderSpec{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OpenStackProviderSpec",
			APIVersion: "cloudcredential.openshift.io/v1",
		},
	}

	return &credsv1.CredentialsRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CredentialsRequest",
			APIVersion: "cloudcredential.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manila-csi-credentials-request",
			Namespace: "openshift-cloud-credential-operator",
		},
		Spec: credsv1.CredentialsRequestSpec{
			SecretRef: corev1.ObjectReference{
				Name:      "installer-cloud-credentials",
				Namespace: "openshift-manila-csi-driver",
			},
			ProviderSpec: &runtime.RawExtension{
				Object: openstackProvSpec,
			},
		},
	}
}
