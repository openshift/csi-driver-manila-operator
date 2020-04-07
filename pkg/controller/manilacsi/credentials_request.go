package manilacsi

import (
	"context"

	"github.com/go-logr/logr"
	credsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaCSI) handleCredentialsRequest(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Credentials Request")

	// Define a new Credential Request object
	creq := generateCredentialsRequest()

	// Check if this Credential Request already exists
	found := &credsv1.CredentialsRequest{}
	// TODO(mfedosin): figure out why we always get NotFound here
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: creq.Name, Namespace: creq.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new CredentialsRequest", "CredentialsRequest.Namespace", creq.Namespace, "CredentialsRequest.Name", creq.Name)
		err = r.client.Create(context.TODO(), creq)
		if err != nil {
			// // TODO(mfedosin): Uncomment this return when cloud credentials fetching is fixed
			// return err
		}

		// Credential Request created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Credential Request already exists - don't requeue
	reqLogger.Info("Skip reconcile: CredentialsRequest already exists", "CredentialsRequest.Namespace", found.Namespace, "CredentialsRequest.Name", found.Name)
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
				Namespace: "manila-csi",
			},
			ProviderSpec: &runtime.RawExtension{
				Object: openstackProvSpec,
			},
		},
	}
}
