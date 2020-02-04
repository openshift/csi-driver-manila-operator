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
	nfsNodePluginManifest = `kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: csi-nodeplugin-nfsplugin
spec:
  selector:
    matchLabels:
      app: csi-nodeplugin-nfsplugin
  template:
    metadata:
      labels:
        app: csi-nodeplugin-nfsplugin
    spec:
      containers:
        - name: nfs
          securityContext:
            privileged: true
            capabilities:
              add: ["SYS_ADMIN"]
            allowPrivilegeEscalation: true
          image: quay.io/k8scsi/nfsplugin:canary
          args:
            - "--nodeid=$(NODE_ID)"
            - "--endpoint=unix://plugin/csi.sock"
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          imagePullPolicy: IfNotPresent
          volumeMounts:
            - name: plugin-dir
              mountPath: /plugin
            - name: pods-mount-dir
              mountPath: /var/lib/kubelet/pods
              mountPropagation: Bidirectional
      volumes:
        - name: plugin-dir
          hostPath:
            path: /var/lib/kubelet/plugins/csi-nfsplugin
            type: DirectoryOrCreate
        - name: pods-mount-dir
          hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
`
)

func (r *ReconcileManilaCSI) handleNFSNodePluginDaemonSet(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	// Define a new Pod object
	ds := &appsv1.DaemonSet{}

	dec := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(nfsNodePluginManifest)), 1000)
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
