package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gophercloud/utils/openstack/clientconfig"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	cloudsSecretKey     = "clouds.yaml"
	installerSecretName = "installer-cloud-credentials"
	driverSecretName    = "csi-manila-secrets"
	secretNamespace     = "openshift-manila-csi-driver"
	cloudName           = "openstack"
)

// createDriverCredentialsSecret converts the installer secret, if it is available, into the driver secret
func (r *ReconcileManilaDriver) createDriverCredentialsSecret(instance *maniladriverv1alpha1.ManilaDriver, cloudConfig clientconfig.Cloud, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Credentials")

	cert, err := r.getCloudProviderCert()
	if err != nil {
		return err
	}

	if cert != "" {
		cloudConfig.CACertFile = "/usr/share/pki/ca-trust-source/cloud-provider-ca-bundle.pem"
	}

	secret := generateSecret(cloudConfig)

	if err := annotator.SetLastAppliedAnnotation(secret); err != nil {
		return err
	}

	found := &corev1.Secret{}
	err = r.apiReader.Get(context.TODO(), types.NamespacedName{Name: driverSecretName, Namespace: secretNamespace}, found)
	if err == nil {
		// Check if we need to update the object
		equal, err := compareLastAppliedAnnotations(found, secret)
		if err != nil {
			return err
		}

		if !equal {
			reqLogger.Info("Updating Secret with new changes", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
			err = r.client.Update(context.TODO(), secret)
			if err != nil {
				return err
			}
		} else {
			// Service already exists - don't requeue
			reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
		}

		return nil
	}

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Convert the installer secret into the driver secret
	reqLogger.Info("Creating a new Secret", "Secret.Namespace", secretNamespace, "Secret.Name", driverSecretName)
	err = r.client.Create(context.TODO(), secret)
	if err != nil {
		return err
	}

	// Secret created successfully - don't requeue
	return nil
}

func generateSecret(cloud clientconfig.Cloud) *corev1.Secret {
	data := make(map[string][]byte)

	if cloud.AuthInfo.AuthURL != "" {
		data["os-authURL"] = []byte(cloud.AuthInfo.AuthURL)
	}
	if cloud.RegionName != "" {
		data["os-region"] = []byte(cloud.RegionName)
	}
	if cloud.AuthInfo.UserID != "" {
		data["os-userID"] = []byte(cloud.AuthInfo.UserID)
	} else if cloud.AuthInfo.Username != "" {
		data["os-userName"] = []byte(cloud.AuthInfo.Username)
	}
	if cloud.AuthInfo.Password != "" {
		data["os-password"] = []byte(cloud.AuthInfo.Password)
	}
	if cloud.AuthInfo.ProjectID != "" {
		data["os-projectID"] = []byte(cloud.AuthInfo.ProjectID)
	} else if cloud.AuthInfo.ProjectName != "" {
		data["os-projectName"] = []byte(cloud.AuthInfo.ProjectName)
	}
	if cloud.AuthInfo.DomainID != "" {
		data["os-domainID"] = []byte(cloud.AuthInfo.DomainID)
	} else if cloud.AuthInfo.DomainName != "" {
		data["os-domainName"] = []byte(cloud.AuthInfo.DomainName)
	}
	if cloud.AuthInfo.ProjectDomainID != "" {
		data["os-projectDomainID"] = []byte(cloud.AuthInfo.ProjectDomainID)
	} else if cloud.AuthInfo.ProjectDomainName != "" {
		data["os-projectDomainName"] = []byte(cloud.AuthInfo.ProjectDomainName)
	}
	if cloud.AuthInfo.UserDomainID != "" {
		data["os-userDomainID"] = []byte(cloud.AuthInfo.UserDomainID)
		data["os-domainID"] = []byte(cloud.AuthInfo.UserDomainID)
	} else if cloud.AuthInfo.UserDomainName != "" {
		data["os-userDomainName"] = []byte(cloud.AuthInfo.UserDomainName)
		data["os-domainName"] = []byte(cloud.AuthInfo.UserDomainName)
	}
	if cloud.CACertFile != "" {
		data["os-certAuthorityPath"] = []byte(cloud.CACertFile)
	}

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      driverSecretName,
			Namespace: secretNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	return &secret
}
