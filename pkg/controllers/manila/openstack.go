package manila

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/openshift/csi-driver-manila-operator/pkg/util"
	"github.com/openshift/csi-driver-manila-operator/pkg/version"
)

func GetShareTypes() ([]sharetypes.ShareType, error) {
	opts := new(clientconfig.ClientOpts)
	opts.Cloud = util.CloudName

	client, err := clientconfig.NewServiceClient("sharev2", opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared filesystem client: %w", err)
	}

	// we represent version using commits since we don't tag releases
	ua := gophercloud.UserAgent{}
	ua.Prepend(fmt.Sprintf("csi-driver-manila-operator/%s", version.Get().GitCommit))
	client.UserAgent = ua

	allPages, err := sharetypes.List(client, &sharetypes.ListOpts{}).AllPages()
	if err != nil {
		return nil, fmt.Errorf("cannot list available share types: %w", err)
	}

	return sharetypes.ExtractShareTypes(allPages)
}
