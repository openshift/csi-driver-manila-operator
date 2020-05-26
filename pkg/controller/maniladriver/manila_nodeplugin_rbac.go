package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	labelsManilaNodePlugin = map[string]string{
		"app":       "openstack-manila-csi",
		"component": "nodeplugin",
	}
)

func (r *ReconcileManilaDriver) handleManilaNodePluginRBAC(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Node Plugin RBAC resources")

	// Manila Node Plugin Service Account
	err := r.handleManilaNodePluginServiceAccount(instance, reqLogger)
	if err != nil {
		return err
	}

	// Manila Node Plugin Cluster Role
	err = r.handleManilaNodePluginClusterRole(instance, reqLogger)
	if err != nil {
		return err
	}

	// Manila Node Plugin Cluster Role Binding
	err = r.handleManilaNodePluginClusterRoleBinding(instance, reqLogger)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileManilaDriver) handleManilaNodePluginServiceAccount(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Node Plugin Service Account")

	// Define a new ServiceAccount object
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-manila-csi-nodeplugin",
			Namespace: "openshift-manila-csi-driver",
			Labels:    labelsManilaNodePlugin,
		},
	}

	if err := annotator.SetLastAppliedAnnotation(sa); err != nil {
		return err
	}

	// Check if this ServiceAccount already exists
	found := &corev1.ServiceAccount{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
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

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, sa)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating ServiceAccount with new changes", "ServiceAccount.Namespace", found.Namespace, "ServiceAccount.Name", found.Name)
		err = r.client.Update(context.TODO(), sa)
		if err != nil {
			return err
		}
	} else {
		// ServiceAccount already exists - don't requeue
		reqLogger.Info("Skip reconcile: ServiceAccount already exists", "ServiceAccount.Namespace", found.Namespace, "ServiceAccount.Name", found.Name)
	}

	return nil
}

func (r *ReconcileManilaDriver) handleManilaNodePluginClusterRole(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Node Plugin Cluster Role")

	// Define a new ClusterRole object
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "openstack-manila-csi-nodeplugin",
			Labels: labelsManilaNodePlugin,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "update"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "update"},
			},
		},
	}

	if err := annotator.SetLastAppliedAnnotation(cr); err != nil {
		return err
	}

	// Check if this ClusterRole already exists
	found := &rbacv1.ClusterRole{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: ""}, found)
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

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, cr)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating ClusterRole with new changes", "ClusterRole.Name", found.Name)
		err = r.client.Update(context.TODO(), cr)
		if err != nil {
			return err
		}
	} else {
		// ClusterRole already exists - don't requeue
		reqLogger.Info("Skip reconcile: ClusterRole already exists", "ClusterRole.Name", found.Name)
	}

	return nil
}

func (r *ReconcileManilaDriver) handleManilaNodePluginClusterRoleBinding(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Node Plugin Cluster Role Binding")

	// Define a new ClusterRoleBinding object
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "openstack-manila-csi-nodeplugin",
			Labels: labelsManilaNodePlugin,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "openstack-manila-csi-nodeplugin",
				Namespace: "openshift-manila-csi-driver",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "openstack-manila-csi-nodeplugin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	if err := annotator.SetLastAppliedAnnotation(crb); err != nil {
		return err
	}

	// Check if this ClusterRoleBinding already exists
	found := &rbacv1.ClusterRoleBinding{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: crb.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ClusterRoleBinding", "ClusterRoleBinding.Name", crb.Name)
		err = r.client.Create(context.TODO(), crb)
		if err != nil {
			return err
		}

		// ClusterRoleBinding created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, crb)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating ClusterRoleBinding with new changes", "ClusterRoleBinding.Name", found.Name)
		err = r.client.Update(context.TODO(), crb)
		if err != nil {
			return err
		}
	} else {
		// ClusterRoleBinding already exists - don't requeue
		reqLogger.Info("Skip reconcile: ClusterRoleBinding already exists", "ClusterRoleBinding.Name", found.Name)
	}

	return nil
}
