package manilacsi

import (
	"bytes"
	"context"

	"github.com/go-logr/logr"
	securityv1 "github.com/openshift/api/security/v1"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	manilaSCCManifest = `apiVersion: security.openshift.io/v1
kind: SecurityContextConstraints
metadata:
  name: csi-driver-manila-operator
allowPrivilegedContainer: true
allowPrivilegeEscalation: true
allowHostDirVolumePlugin: true
allowedCapabilities:
- SYS_ADMIN
allowHostIPC: true
allowHostNetwork: true
allowHostPID: false
allowHostPorts: false
runAsUser:
  type: RunAsAny
seLinuxContext:
  type: RunAsAny
fsGroup:
  type: RunAsAny
supplementalGroups:
  type: RunAsAny
users:
- system:serviceaccount:manila-csi:csi-nodeplugin
- system:serviceaccount:manila-csi:openstack-manila-csi-controllerplugin
- system:serviceaccount:manila-csi:openstack-manila-csi-nodeplugin
groups: []
volumes:
- configMap
- downwardAPI
- emptyDir
- hostPath
- nfs
- persistentVolumeClaim
- projected
- secret
`
)

func (r *ReconcileManilaCSI) handleSecurityContextConstraints(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Security Context Constraints")

	// Define a new Security Context Constraints object
	scc, err := generateSecurityContextConstraints()
	if err != nil {
		return err
	}

	// Check if this Security Context Constraints already exists
	found := &securityv1.SecurityContextConstraints{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: scc.Name, Namespace: ""}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new SecurityContextConstraints", "SecurityContextConstraints.Name", scc.Name)
		err = r.client.Create(context.TODO(), scc)
		if err != nil {
			return err
		}

		// Security Context Constraints created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Security Context Constraints already exists - don't requeue
	reqLogger.Info("Skip reconcile: SecurityContextConstraints already exists", "SecurityContextConstraints.Name", found.Name)
	return nil
}

func generateSecurityContextConstraints() (*securityv1.SecurityContextConstraints, error) {
	scc := &securityv1.SecurityContextConstraints{}

	dec := k8sYaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manilaSCCManifest)), 1000)
	if err := dec.Decode(scc); err != nil {
		return nil, err
	}

	return scc, nil
}
