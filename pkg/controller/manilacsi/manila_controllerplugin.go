package manilacsi

import (
	"bytes"
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	manilaControllerPluginManifest = `kind: StatefulSet
apiVersion: apps/v1
metadata:
  name: openstack-manila-csi-controllerplugin
  labels:
    app: openstack-manila-csi
    component: controllerplugin
spec:
  serviceName: openstack-manila-csi-controllerplugin
  replicas: 1
  selector:
    matchLabels:
      app: openstack-manila-csi
      component: controllerplugin
  template:
    metadata:
      labels:
        app: openstack-manila-csi
        component: controllerplugin
    spec:
      serviceAccountName: openstack-manila-csi-controllerplugin
      containers:
        - name: provisioner
          image: "quay.io/k8scsi/csi-provisioner:v1.4.0"
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
          env:
            - name: ADDRESS
              value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock"
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: plugin-dir
              mountPath: /var/lib/kubelet/plugins/manila.csi.openstack.org
        - name: snapshotter
          image: "quay.io/k8scsi/csi-snapshotter:v1.2.2"
          args:
            - "--v=5"
            - "--csi-address=$(ADDRESS)"
          env:
            - name: ADDRESS
              value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock"
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: plugin-dir
              mountPath: /var/lib/kubelet/plugins/manila.csi.openstack.org
        - name: nodeplugin
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: "manila-csi-plugin:latest"
          args:
            - "--v=5"
            - "--nodeid=$(NODE_ID)"
            - "--endpoint=$(CSI_ENDPOINT)"
            - "--drivername=$(DRIVER_NAME)"
            - "--share-protocol-selector=$(MANILA_SHARE_PROTO)"
            - "--fwdendpoint=$(FWD_CSI_ENDPOINT)"
          env:
            - name: DRIVER_NAME
              value: manila.csi.openstack.org
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: CSI_ENDPOINT
              value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock"
            - name: FWD_CSI_ENDPOINT
              value: "unix:///var/lib/kubelet/plugins/csi-nfsplugin/csi.sock"
            - name: MANILA_SHARE_PROTO
              value: "NFS"
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: plugin-dir
              mountPath: /var/lib/kubelet/plugins/manila.csi.openstack.org
            - name: fwd-plugin-dir
              mountPath: /var/lib/kubelet/plugins/csi-nfsplugin
            - name: pod-mounts
              mountPath: /var/lib/kubelet/pods
              mountPropagation: Bidirectional
      volumes:
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/manila.csi.openstack.org
            type: DirectoryOrCreate
        - name: fwd-plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-nfsplugin
            type: Directory
        - name: pod-mounts
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
`
)

func (r *ReconcileManilaCSI) handleManilaControllerPluginStatefulSet(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Controller Plugin StatefulSet")

	// Define a new StatefulSet object
	ss := &appsv1.StatefulSet{}

	dec := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manilaControllerPluginManifest)), 1000)
	if err := dec.Decode(&ss); err != nil {
		return err
	}

	// Set ManilaCSI instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ss, r.scheme); err != nil {
		return err
	}

	// Check if this StatefulSet already exists
	found := &appsv1.StatefulSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", ss.Namespace, "StatefulSet.Name", ss.Name)
		err = r.client.Create(context.TODO(), ss)
		if err != nil {
			return err
		}

		// StatefulSet created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// DaemonSet already exists - don't requeue
	reqLogger.Info("Skip reconcile: StatefulSet already exists", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
	return nil
}
