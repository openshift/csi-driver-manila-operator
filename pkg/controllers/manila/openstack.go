package manila

import (
	"fmt"
	"io/ioutil"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/openshift/csi-driver-manila-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
	corelisters "k8s.io/client-go/listers/core/v1"
	"sigs.k8s.io/yaml"
)

type openStackClient struct {
	cloud           *clientconfig.Cloud
	configMapLister corelisters.ConfigMapLister
}

func NewOpenStackClient(
	cloudConfigFilename string,
	informers v1helpers.KubeInformersForNamespaces,
) (*openStackClient, error) {
	cloud, err := getCloudFromFile(cloudConfigFilename)
	if err != nil {
		return nil, err
	}
	return &openStackClient{
		cloud:           cloud,
		configMapLister: informers.InformersFor(util.CloudConfigNamespace).Core().V1().ConfigMaps().Lister(),
	}, nil
}

func (o *openStackClient) GetShareTypes() ([]sharetypes.ShareType, error) {
	clientOpts := new(clientconfig.ClientOpts)

	if o.cloud.AuthInfo != nil {
		clientOpts.AuthInfo = o.cloud.AuthInfo
		clientOpts.AuthType = o.cloud.AuthType
		clientOpts.Cloud = o.cloud.Cloud
		clientOpts.RegionName = o.cloud.RegionName
	}

	opts, err := clientconfig.AuthOptions(clientOpts)
	if err != nil {
		return nil, err
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	err = openstack.Authenticate(provider, *opts)
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewSharedFileSystemV2(provider, gophercloud.EndpointOpts{
		Region: clientOpts.RegionName,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := sharetypes.List(client, &sharetypes.ListOpts{}).AllPages()
	if err != nil {
		return nil, err
	}

	return sharetypes.ExtractShareTypes(allPages)
}

func getCloudFromFile(filename string) (*clientconfig.Cloud, error) {
	cloudConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var clouds clientconfig.Clouds
	err = yaml.Unmarshal(cloudConfig, &clouds)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal clouds credentials from %s: %v", filename, err)
	}

	cfg, ok := clouds.Clouds[util.CloudName]
	if !ok {
		return nil, fmt.Errorf("could not find cloud named %q in credential file %s", util.CloudName, filename)
	}
	return &cfg, nil
}
