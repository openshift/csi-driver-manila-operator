package manilacsi

import (
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaCSI) handleNFSNodePluginRBAC(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling NFS Node Plugin RBAC resources")

	// NFS Node Plugin Service Account
	err := r.handleNFSNodePluginServiceAccount(instance, reqLogger)
	if err != nil {
		return err
	}

	// NFS Node Plugin Cluster Role
	err = r.handleNFSNodePluginClusterRole(instance, reqLogger)
	if err != nil {
		return err
	}

	// NFS Node Plugin Cluster Role Binding
	err = r.handleNFSNodePluginClusterRole(instance, reqLogger)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManilaCSI) handleNFSNodePluginServiceAccount(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling NFS Node Plugin Service Account")

	// Define a new ServiceAccount object
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-nodeplugin",
			Namespace: "default",
		},
	}

	// Check if this ServiceAccount already exists
	found := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
		err = r.client.Create(context.TODO(), sa)
		if err != nil {
			return err
		}

		// ServiceAccount created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// ServiceAccount already exists - don't requeue
	reqLogger.Info("Skip reconcile: ServiceAccount already exists", "ServiceAccount.Namespace", found.Namespace, "ServiceAccount.Name", found.Name)
	return nil
}

func (r *ReconcileManilaCSI) handleNFSNodePluginClusterRole(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling NFS Node Plugin Cluster Role")

	// Define a new ClusterRole object
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "csi-nodeplugin",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"volumeattachments"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
		},
	}

	// Check if this ClusterRole already exists
	found := &rbacv1.ClusterRole{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ClusterRole", "ClusterRole.Name", cr.Name)
		err = r.client.Create(context.TODO(), cr)
		if err != nil {
			return err
		}

		// ClusterRole created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// ClusterRole already exists - don't requeue
	reqLogger.Info("Skip reconcile: ClusterRole already exists", "ClusterRole.Name", found.Name)
	return nil
}

func (r *ReconcileManilaCSI) handleNFSNodePluginClusterRoleBinding(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling NFS Node Plugin Cluster Role Binding")

	// Define a new ClusterRoleBinding object
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "csi-nodeplugin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "csi-nodeplugin",
				Namespace: "default",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "csi-nodeplugin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	// Check if this ClusterRoleBinding already exists
	found := &rbacv1.ClusterRoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crb.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ClusterRole", "ClusterRoleBinding.Name", crb.Name)
		err = r.client.Create(context.TODO(), crb)
		if err != nil {
			return err
		}

		// ClusterRoleBinding created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// ClusterRoleBinding already exists - don't requeue
	reqLogger.Info("Skip reconcile: ClusterRoleBinding already exists", "ClusterRoleBinding.Name", found.Name)
	return nil
}
