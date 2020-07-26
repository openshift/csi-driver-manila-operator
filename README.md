# Manila CSI driver operator

An operator to deploy the [Manil CSI driver](https://github.com/openshift/cloud-provider-openstack/tree/master/pkg/csi/manila) in OpenShift.

## Design

The operator is based on [openshift/library-go](https://github.com/openshift/library-go). It manages `ManilaDriver` instance named `cluster` and runs several controllers in parallel:

* `manilaController`: Talks to OpenStack API and checks if Manila service is provided.
  * If Manila service is found: 
    * It starts `manilaControllerSet`: Runs `csidriverset.Controller` that installs the Manila CSI driver itself.
    * It starts `nfsController`: Runs `csidriverset.Controller` that installs NFS CSI driver itself.
    * It creates `StorageClass` for each share type reported by Manila and periodically syncs them at least once per minute, in case a new share type appears in Manila.
    * It never removes StorageClass if Manila service or share type disappears. It may be temporary OpenStack ir Manila re-configuration hiccup.
  * If there is no Manila service, it marks the `ManilaDriver` instance with `PrereqSatisfied: False` condition. It does not stop any CSI drivers started when Manila service was present! This allows pod to at least unmount their volumes. 
* `secretSyncController`: Syncs Secret provided by cloud-credentials-operator into a new Secret that is used by the CSI drivers. The drivers need OpenStack credentials in different format than provided by cloud-credentials-operator.

## Usage

TODO
