package manilacsi

import (
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_manilacsi")

// Add creates a new ManilaCSI Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManilaCSI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("manilacsi-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ManilaCSI
	err = c.Watch(&source.Kind{Type: &manilacsiv1alpha1.ManilaCSI{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch owned objects
	watchOwnedObjects := []runtime.Object{
		&appsv1.StatefulSet{},
		&appsv1.DaemonSet{},
		&corev1.Service{},
		&storagev1beta1.CSIDriver{},
		&corev1.ServiceAccount{},
		&rbacv1.ClusterRole{},
		&rbacv1.ClusterRoleBinding{},
		&rbacv1.Role{},
		&rbacv1.RoleBinding{},
	}

	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &manilacsiv1alpha1.ManilaCSI{},
	}

	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			return err
		}
	}

	return nil
}

// blank assignment to verify that ReconcileManilaCSI implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManilaCSI{}

// ReconcileManilaCSI reconciles a ManilaCSI object
type ReconcileManilaCSI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManilaCSI object and makes changes based on the state read
// and what is in the ManilaCSI.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileManilaCSI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManilaCSI")

	// Fetch the ManilaCSI instance
	instance := &manilacsiv1alpha1.ManilaCSI{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	// Manage objects created by the operator
	return r.handleManilaCSIDeployment(instance, reqLogger)
}

// Manage the Objects created by the Operator.
func (r *ReconcileManilaCSI) handleManilaCSIDeployment(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) (reconcile.Result, error) {
	reqLogger.Info("Reconciling ManilaCSI Deployment Objects")

	// NFS Node Plugin RBAC
	err := r.handleNFSNodePluginRBAC(instance, reqLogger)
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
