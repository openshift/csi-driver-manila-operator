package secret

import (
	"context"
	"fmt"

	"github.com/gophercloud/utils/openstack/clientconfig"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/csi-driver-manila-operator/pkg/util"
	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog"
)

// This Controller translates Secret provided by cloud-credential-operator into
// format required by the CSI driver.
type Controller struct {
	operatorClient v1helpers.OperatorClient
	kubeClient     kubernetes.Interface
	secretLister   corelisters.SecretLister
	eventRecorder  events.Recorder
}

const (
	// Name of key with clouds.yaml in Secret provided by cloud-credentials-operator.
	cloudSecretKey = "clouds.yaml"
	// Name of OpenStack in clouds.yaml
	cloudName = "openstack"
)

func NewController(
	operatorClient v1helpers.OperatorClient,
	kubeClient kubernetes.Interface,
	informers v1helpers.KubeInformersForNamespaces,
	eventRecorder events.Recorder) factory.Controller {

	// Watch config-map
	// Produce secrets in the operator namespace
	secretInformer := informers.InformersFor(util.OperatorNamespace)
	c := &Controller{
		operatorClient: operatorClient,
		kubeClient:     kubeClient,
		secretLister:   secretInformer.Core().V1().Secrets().Lister(),
		eventRecorder:  eventRecorder,
	}
	return factory.New().WithSync(c.sync).WithSyncDegradedOnError(operatorClient).WithInformers(
		operatorClient.Informer(),
		secretInformer.Core().V1().Secrets().Informer(),
	).ToController("SecretSync", eventRecorder)
}

func (c *Controller) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	opSpec, _, _, err := c.operatorClient.GetOperatorState()
	if err != nil {
		return err
	}
	if opSpec.ManagementState != operatorv1.Managed {
		return nil
	}

	cloudSecret, err := c.secretLister.Secrets(util.OperatorNamespace).Get(util.CloudCredentialSecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			// TODO: report error after some while?
			klog.V(2).Infof("Waiting for secret %s from cloud-credentials-operator", util.CloudCredentialSecretName)
			return nil
		}
		return err
	}

	driverSecret, err := c.translateSecret(cloudSecret)
	if err != nil {
		return err
	}

	_, _, err = resourceapply.ApplySecret(c.kubeClient.CoreV1(), c.eventRecorder, driverSecret)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) translateSecret(cloudSecret *v1.Secret) (*v1.Secret, error) {
	content, ok := cloudSecret.Data[cloudSecretKey]
	if !ok {
		return nil, fmt.Errorf("OpenStack credentials secret %s did not contain key %v", util.CloudCredentialSecretName, cloudSecretKey)
	}

	var clouds clientconfig.Clouds
	err := yaml.Unmarshal(content, &clouds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal clouds credentials stored in secret %s: %v", util.CloudCredentialSecretName, err)
	}

	cloud, ok := clouds.Clouds[cloudName]
	if !ok {
		return nil, fmt.Errorf("failed to parse clouds credentials stored in secret %s: cloud %s not found", util.CloudCredentialSecretName, cloudName)
	}

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

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.ManilaSecretName,
			Namespace: util.OperatorNamespace,
		},
		Type: v1.SecretTypeOpaque,
		Data: data,
	}

	return &secret, nil
}
