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
	manilaNodePluginManifest = `kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: openstack-manila-csi-nodeplugin
  labels:
    app: openstack-manila-csi
    component: nodeplugin
spec:
  selector:
    matchLabels:
      app: openstack-manila-csi
      component: nodeplugin
  template:
    metadata:
      labels:
        app: openstack-manila-csi
        component: nodeplugin
    spec:
      serviceAccountName: openstack-manila-csi-nodeplugin
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      containers:
        - name: registrar
          image: "quay.io/k8scsi/csi-node-driver-registrar:v1.1.0"
          args:
            - "--v=5"
            - "--csi-address=/csi/csi.sock"
            - "--kubelet-registration-path=/var/lib/kubelet/plugins/manila.csi.openstack.org/csi.sock"
          lifecycle:
            preStop:
              exec:
                  command: [
                    "/bin/sh", "-c",
                    'rm -rf /registration/manila.csi.openstack.org
                  /registration/manila.csi.openstack.org-reg.sock'
                  ]
          env:
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: plugin-dir
              mountPath: /csi
            - name: registration-dir
              mountPath: /registration
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
              value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi.sock"
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
      volumes:
        - name: registration-dir
          hostPath:
            path: /var/lib/kubelet/plugins_registry
            type: Directory
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/manila.csi.openstack.org
            type: DirectoryOrCreate
        - name: fwd-plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-nfsplugin
            type: Directory
`
)

func (r *ReconcileManilaCSI) handleManilaNodePluginDaemonSet(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	// Define a new Pod object
	ds := &appsv1.DaemonSet{}

	dec := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manilaNodePluginManifest)), 1000)
	if err := dec.Decode(&ds); err != nil {
		return err
	}

	// Set ManilaCSI instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ds, r.scheme); err != nil {
		return err
	}

	// Check if this DaemonSet already exists
	found := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: ds.Name, Namespace: ds.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new DaemonSet", "DaemonSet.Namespace", ds.Namespace, "DaemonSet.Name", ds.Name)
		err = r.client.Create(context.TODO(), ds)
		if err != nil {
			return err
		}

		// DaemonSet created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// DaemonSet already exists - don't requeue
	reqLogger.Info("Skip reconcile: DaemonSet already exists", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
	return nil
}
