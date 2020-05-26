package maniladriver

import (
	"context"
	"fmt"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/sharetypes"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/nsf/jsondiff"
	securityv1 "github.com/openshift/api/security/v1"
	credsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	lastAppliedAnnotationName = "manila.csi.openshift.io/last-applied"
)

var log = logf.Log.WithName("controller_maniladriver")

var annotator = patch.NewAnnotator("manila.csi.openshift.io/last-applied")

func compareLastAppliedAnnotations(currentObject, modifiedObject runtime.Object) (bool, error) {
	currentAnnotation, err := annotator.GetOriginalConfiguration(currentObject)
	if err != nil {
		return false, err
	}

	modifiedAnnotation, err := annotator.GetOriginalConfiguration(modifiedObject)
	if err != nil {
		return false, err
	}

	opts := jsondiff.DefaultJSONOptions()
	diff, _ := jsondiff.Compare(currentAnnotation, modifiedAnnotation, &opts)

	return diff == jsondiff.FullMatch, nil
}

// Add creates a new ManilaDriver Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManilaDriver{client: mgr.GetClient(), scheme: mgr.GetScheme(), apiReader: mgr.GetAPIReader()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("maniladriver-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ManilaDriver
	err = c.Watch(&source.Kind{Type: &maniladriverv1alpha1.ManilaDriver{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch owned objects
	watchOwnedObjects := []runtime.Object{
		&appsv1.StatefulSet{},
		&appsv1.DaemonSet{},
		&corev1.Namespace{},
		&corev1.Secret{},
		&corev1.Service{},
		&storagev1beta1.CSIDriver{},
		&storagev1.StorageClass{},
		&corev1.ServiceAccount{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
		&credsv1.CredentialsRequest{},
		&securityv1.SecurityContextConstraints{},
	}

	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &maniladriverv1alpha1.ManilaDriver{},
	}

	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileManilaDriver implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManilaDriver{}

// ReconcileManilaDriver reconciles a ManilaDriver object
type ReconcileManilaDriver struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	apiReader client.Reader
}

// Reconcile reads that state of the cluster for a ManilaDriver object and makes changes based on the state read
// and what is in the ManilaDriver.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileManilaDriver) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManilaDriver")

	// Fetch the ManilaDriver instance
	instance := &maniladriverv1alpha1.ManilaDriver{}
	err := r.apiReader.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get %v: %v", request.NamespacedName, err)
		return reconcile.Result{}, err
	}

	// Manila Driver Namespace
	err = r.handleManilaDriverNamespace(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Credentials Request
	err = r.handleCredentialsRequest(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Get the cloud credentials
	cloud, err := r.getCloudFromSecret()
	if err != nil {
		// It can take a while before the secret is created
		if errors.IsNotFound(err) {
			reqLogger.Info(fmt.Sprintf("No %v secret was found in %v namespace. Retrying...", installerSecretName, secretNamespace))
			return reconcile.Result{
				RequeueAfter: 10,
			}, nil
		}
		return reconcile.Result{}, err
	}

	// Driver Secret
	err = r.createDriverCredentialsSecret(instance, cloud, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	reqLogger.Info("Fetching Manila Share Types")
	shareTypes, err := r.getManilaShareTypes(cloud, reqLogger)
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault404); !ok {
			return reconcile.Result{}, err
		}
		reqLogger.Info("OpenStack Manila is not available in the cloud")
		return reconcile.Result{}, nil
	}

	// StorageClasses
	err = r.handleManilaStorageClasses(instance, shareTypes, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manage objects created by the operator
	return r.handleManilariverDeployment(instance, reqLogger)
}

// Manage the Objects created by the Operator.
func (r *ReconcileManilaDriver) handleManilariverDeployment(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) (reconcile.Result, error) {
	reqLogger.Info("Reconciling ManilaDriver Deployment Objects")

	// Security Context Constraints
	err := r.handleSecurityContextConstraints(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// NFS Node Plugin RBAC
	err = r.handleNFSNodePluginRBAC(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// NFS Node Plugin DaemonSet
	err = r.handleNFSNodePluginDaemonSet(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// CSIDriver
	err = r.handleManilaCSIDriver(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manila Controller Plugin RBAC
	err = r.handleManilaControllerPluginRBAC(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manila Controller Plugin Service
	err = r.handleManilaControllerPluginService(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manila Controller Plugin StatefulSet
	err = r.handleManilaControllerPluginStatefulSet(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manila Node Plugin RBAC
	err = r.handleManilaNodePluginRBAC(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Manila Node Plugin DaemonSet
	err = r.handleManilaNodePluginDaemonSet(instance, reqLogger)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// getManilaShareTypes returns all available share types
func (r *ReconcileManilaDriver) getManilaShareTypes(cloud clientconfig.Cloud, reqLogger logr.Logger) ([]sharetypes.ShareType, error) {
	clientOpts := new(clientconfig.ClientOpts)

	if cloud.AuthInfo != nil {
		clientOpts.AuthInfo = cloud.AuthInfo
		clientOpts.AuthType = cloud.AuthType
		clientOpts.Cloud = cloud.Cloud
		clientOpts.RegionName = cloud.RegionName
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

// getCloudFromSecret extract a Cloud from the given namespace:secretName
func (r *ReconcileManilaDriver) getCloudFromSecret() (clientconfig.Cloud, error) {
	ctx := context.TODO()
	emptyCloud := clientconfig.Cloud{}

	secret := &corev1.Secret{}
	err := r.apiReader.Get(ctx, types.NamespacedName{
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
