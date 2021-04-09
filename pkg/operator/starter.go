package operator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openshift/csi-driver-manila-operator/pkg/controllers/manila"
	"github.com/openshift/csi-driver-manila-operator/pkg/controllers/secret"
	"github.com/openshift/csi-driver-manila-operator/pkg/generated"
	"github.com/openshift/csi-driver-manila-operator/pkg/util"
	"github.com/openshift/library-go/pkg/operator/resourcesynccontroller"
	"k8s.io/client-go/dynamic"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	operatorapi "github.com/openshift/api/operator/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformers "github.com/openshift/client-go/config/informers/externalversions"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	csicontrollerset "github.com/openshift/library-go/pkg/operator/csi/csicontrollerset"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivercontrollerservicecontroller"
	"github.com/openshift/library-go/pkg/operator/csi/csidrivernodeservicecontroller"

	// csidrivercontroller "github.com/openshift/library-go/pkg/operator/csi/csidrivercontroller"
	goc "github.com/openshift/library-go/pkg/operator/genericoperatorclient"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

const (
	operandName  = "manila-csi-driver"
	operatorName = "manila-csi-driver-operator"

	nfsImageEnvName = "NFS_DRIVER_IMAGE"

	resync = 20 * time.Minute
)

func RunOperator(ctx context.Context, controllerConfig *controllercmd.ControllerContext) error {
	kubeClient := kubeclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	kubeInformersForNamespaces := v1helpers.NewKubeInformersForNamespaces(kubeClient, util.OperatorNamespace, util.OperandNamespace, util.CloudConfigNamespace, "")

	configClient := configclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	configInformers := configinformers.NewSharedInformerFactory(configClient, resync)

	// Create GenericOperatorclient. This is used by controllers created down below
	gvr := operatorapi.SchemeGroupVersion.WithResource("clustercsidrivers")
	operatorClient, dynamicInformers, err := goc.NewClusterScopedOperatorClientWithConfigName(controllerConfig.KubeConfig, gvr, string(operatorapi.ManilaCSIDriver))
	if err != nil {
		return err
	}

	dynamicClient, err := dynamic.NewForConfig(controllerConfig.KubeConfig)
	if err != nil {
		return err
	}

	csiDriverControllerSet := csicontrollerset.NewCSIControllerSet(
		operatorClient,
		controllerConfig.EventRecorder,
	).WithLogLevelController().WithManagementStateController(
		operandName,
		false,
	).WithStaticResourcesController(
		"ManilaDriverStaticResources",
		kubeClient,
		kubeInformersForNamespaces,
		generated.Asset,
		[]string{
			"namespace.yaml",
			"csidriver.yaml",
			"controller_sa.yaml",
			"node_sa.yaml",
			"service.yaml",
			"rbac/snapshotter_binding.yaml",
			"rbac/snapshotter_role.yaml",
			"rbac/provisioner_binding.yaml",
			"rbac/provisioner_role.yaml",
			"rbac/privileged_role.yaml",
			"rbac/controller_privileged_binding.yaml",
			"rbac/node_privileged_binding.yaml",
			"rbac/kube_rbac_proxy_role.yaml",
			"rbac/kube_rbac_proxy_binding.yaml",
			"rbac/prometheus_role.yaml",
			"rbac/prometheus_rolebinding.yaml",
		},
	).WithCSIConfigObserverController(
		"ManilaDriverCSIConfigObserverController",
		configInformers,
	).WithCSIDriverControllerService(
		"ManilaDriverControllerServiceController",
		assetWithNFSDriver,
		"controller.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(util.OperandNamespace),
		nil,
		csidrivercontrollerservicecontroller.WithObservedProxyDeploymentHook(),
	).WithCSIDriverNodeService(
		"ManilaDriverNodeServiceController",
		assetWithNFSDriver,
		"node.yaml",
		kubeClient,
		kubeInformersForNamespaces.InformersFor(util.OperandNamespace),
		csidrivernodeservicecontroller.WithObservedProxyDaemonSetHook(),
	).WithServiceMonitorController(
		"ManilaDriverServiceMonitorController",
		dynamicClient,
		generated.Asset,
		"servicemonitor.yaml",
	).WithExtraInformers(configInformers.Config().V1().Proxies().Informer())

	nfsCSIDriverController := csidrivernodeservicecontroller.NewCSIDriverNodeServiceController(
		"NFSDriverNodeServiceController",
		assetWithNFSDriver("node_nfs.yaml"),
		operatorClient,
		kubeClient,
		kubeInformersForNamespaces.InformersFor(util.OperandNamespace).Apps().V1().DaemonSets(),
		controllerConfig.EventRecorder,
	)

	// sync config map with OpenStack CA certificate to the operand namespace,
	// so the driver can get it as a ConfigMap volume.
	srcConfigMap := resourcesynccontroller.ResourceLocation{
		Namespace: util.CloudConfigNamespace,
		Name:      util.CloudConfigName,
	}
	dstConfigMap := resourcesynccontroller.ResourceLocation{
		Namespace: util.OperandNamespace,
		Name:      util.CloudConfigName,
	}
	certController := resourcesynccontroller.NewResourceSyncController(
		operatorClient,
		kubeInformersForNamespaces,
		kubeClient.CoreV1(),
		kubeClient.CoreV1(),
		controllerConfig.EventRecorder)
	if err := certController.SyncConfigMap(dstConfigMap, srcConfigMap); err != nil {
		return err
	}

	openstackClient, err := manila.NewOpenStackClient(util.CloudConfigFilename, kubeInformersForNamespaces)
	if err != nil {
		return err
	}

	secretSyncController := secret.NewSecretSyncController(
		operatorClient,
		kubeClient,
		kubeInformersForNamespaces,
		resync,
		controllerConfig.EventRecorder)

	manilaController := manila.NewManilaController(
		operatorClient,
		kubeClient,
		kubeInformersForNamespaces,
		openstackClient,
		[]manila.Runnable{
			csiDriverControllerSet,
			nfsCSIDriverController,
			secretSyncController,
			certController,
		},
		controllerConfig.EventRecorder,
	)

	klog.Info("Starting the informers")
	go kubeInformersForNamespaces.Start(ctx.Done())
	go dynamicInformers.Start(ctx.Done())
	go configInformers.Start(ctx.Done())

	klog.Info("Starting controllers")
	go manilaController.Run(ctx, 1)

	<-ctx.Done()

	return fmt.Errorf("stopped")
}

// CSIDriverController can replace only a single driver in driver manifests.
// Manila needs to replace two of them: Manila driver and NFS driver image.
// Let the Manila image be replaced by CSIDriverController and NFS in this
// custom asset loading func.
func assetWithNFSDriver(file string) []byte {
	asset := generated.MustAsset(file)
	nfsImage := os.Getenv(nfsImageEnvName)
	if nfsImage == "" {
		return asset
	}
	return bytes.ReplaceAll(asset, []byte("${NFS_DRIVER_IMAGE}"), []byte(nfsImage))
}
