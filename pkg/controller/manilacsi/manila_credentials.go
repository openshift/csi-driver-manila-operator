package manilacsi

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/gophercloud/utils/openstack/clientconfig"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	cloudsSecretKey     = "clouds.yaml"
	installerSecretName = "installer-cloud-credentials"
	driverSecretName    = "csi-manila-secrets"
	secretNamespace     = "manila-csi"
	cloudName           = "openstack"
)

// createDriverCredentialsSecret converts the installer secret, if it is available, into the driver secret
func (r *ReconcileManilaCSI) createDriverCredentialsSecret(instance *manilacsiv1alpha1.ManilaCSI, cloudConfig clientconfig.Cloud, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Credentials")

	secret := generateSecret(cloudConfig)

	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: driverSecretName, Namespace: secretNamespace}, found)
	if err == nil {
		// Check if we need to update the object
		patchResult, err := patch.DefaultPatchMaker.Calculate(found, secret)
		if err != nil {
			return err
		}

		if !patchResult.IsEmpty() {
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
