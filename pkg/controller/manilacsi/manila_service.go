package manilacsi

import (
	"bytes"
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	manilaServiceManifest = `kind: Service
apiVersion: v1
metadata:
	name: openstack-manila-csi-controllerplugin
	namespace: default
	labels:
	app: openstack-manila-csi
	component: controllerplugin
spec:
	selector:
	app: openstack-manila-csi
	component: controllerplugin
	ports:
	- name: dummy
		port: 12345
`
)

func (r *ReconcileManilaCSI) handleManilaControllerPluginService(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Controller Plugin Service")

	// Define a new Service object
	srv := &corev1.Service{}

	dec := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manilaNodePluginManifest)), 1000)
	if err := dec.Decode(&srv); err != nil {
		return err
	}

	// Set ManilaCSI instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, srv, r.scheme); err != nil {
		return err
	}

	// Check if this Service already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: srv.Name, Namespace: srv.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", srv.Namespace, "Service.Name", srv.Name)
		err = r.client.Create(context.TODO(), srv)
		if err != nil {
			return err
		}

		// Service created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Service already exists - don't requeue
	reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", found.Namespace, "Service.Name", found.Name)
	return nil
}
