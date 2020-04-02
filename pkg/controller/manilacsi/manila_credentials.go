package manilacsi

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gophercloud/utils/openstack/clientconfig"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	cloudsSecretKey     = "clouds.yaml"
	installerSecretName = "installer-cloud-credentials"
	driverSecretName    = "csi-manila-secrets"
	secretNamespace     = "manila-csi"
	cloudName           = "openstack"
)

// createDriverCredentialsSecret converts the installer secret, if it is available, into the driver secret
func (r *ReconcileManilaCSI) createDriverCredentialsSecret(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Credentials")

	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: driverSecretName, Namespace: secretNamespace}, found)
	if err == nil {
		reqLogger.Info("Skip reconcile: Secret already exists", "Secret.Namespace", found.Namespace, "Service.Name", found.Name)
		return nil
	}

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Check if the installer secret, created from the credentials request, exists
	cloudConfig, err := r.getCloudFromSecret()
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No %v secret was found in %v namespace. Skip credentials reconciling", installerSecretName, secretNamespace)
			return nil
		}
		return err
	}

	// Convert the installer secret into the driver secret
	reqLogger.Info("Creating a new Secret", "Secret.Namespace", secretNamespace, "Secret.Name", driverSecretName)
	err = r.client.Create(context.TODO(), generateSecret(cloudConfig))
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

// getCloudFromSecret extract a Cloud from the given namespace:secretName
func (r *ReconcileManilaCSI) getCloudFromSecret() (clientconfig.Cloud, error) {
	ctx := context.TODO()
	emptyCloud := clientconfig.Cloud{}

	secret := &corev1.Secret{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: secretNamespace,
		Name:      installerSecretName,
	}, secret)
	if err != nil {
		return emptyCloud, err
	}

	content, ok := secret.Data[cloudsSecretKey]
	if !ok {
		return emptyCloud, fmt.Errorf("OpenStack credentials secret %v did not contain key %v", installerSecretName, cloudsSecretKey)
	}
	var clouds clientconfig.Clouds
	err = yaml.Unmarshal(content, &clouds)
	if err != nil {
		return emptyCloud, fmt.Errorf("failed to unmarshal clouds credentials stored in secret %v: %v", installerSecretName, err)
	}

	return clouds.Clouds[cloudName], nil
}
