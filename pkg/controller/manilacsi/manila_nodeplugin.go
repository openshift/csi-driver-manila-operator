package manilacsi

import (
	"context"

	manilacsiv1alpha1 "github.com/Fedosin/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaCSI) handleManilaNodePluginDaemonSet(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Node Plugin DaemonSet")

	// Define a new DaemonSet object
	ds := generateManilaNodePluginManifest()

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

func generateManilaNodePluginManifest() *appsv1.DaemonSet {
	trueVar := true

	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	hostPathDirectory := corev1.HostPathDirectory

	labels := map[string]string{
		"app":       "openstack-manila-csi",
		"component": "nodeplugin",
	}

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-manila-csi-nodeplugin",
			Namespace: "manila-csi",
			Labels:    labels,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "openstack-manila-csi-nodeplugin",
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					Containers: []corev1.Container{
						{
							Name:  "registrar",
							Image: "quay.io/openshift/origin-csi-node-driver-registrar:latest",
							Args: []string{
								"--v=5",
								"--csi-address=/csi/csi.sock",
								"--kubelet-registration-path=/var/lib/kubelet/plugins/manila.csi.openstack.org/csi.sock",
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"/bin/sh", "-c",
											"rm -rf /registration/manila.csi.openstack.org /registration/manila.csi.openstack.org-reg.sock",
										},
									},
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "KUBE_NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-dir",
									MountPath: "/csi",
								},
								{
									Name:      "registration-dir",
									MountPath: "/registration",
								},
							},
						},
						{
							Name: "nodeplugin",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueVar,
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
									},
								},
								AllowPrivilegeEscalation: &trueVar,
							},
							Image: "quay.io/openshift/origin-csi-driver-manila:latest",
							Args: []string{
								"--v=5",
								"--nodeid=$(NODE_ID)",
								"--endpoint=$(CSI_ENDPOINT)",
								"--drivername=$(DRIVER_NAME)",
								"--share-protocol-selector=$(MANILA_SHARE_PROTO)",
								"--fwdendpoint=$(FWD_CSI_ENDPOINT)",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "DRIVER_NAME",
									Value: "manila.csi.openstack.org",
								},
								{
									Name: "NODE_ID",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
								{
									Name:  "CSI_ENDPOINT",
									Value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi.sock",
								},
								{
									Name:  "FWD_CSI_ENDPOINT",
									Value: "unix:///var/lib/kubelet/plugins/csi-nfsplugin/csi.sock",
								},
								{
									Name:  "MANILA_SHARE_PROTO",
									Value: "NFS",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-dir",
									MountPath: "/var/lib/kubelet/plugins/manila.csi.openstack.org",
								},
								{
									Name:      "fwd-plugin-dir",
									MountPath: "/var/lib/kubelet/plugins/csi-nfsplugin",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "registration-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins_registry",
									Type: &hostPathDirectory,
								},
							},
						},
						{
							Name: "plugin-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins/manila.csi.openstack.org",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "fwd-plugin-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins/csi-nfsplugin",
									Type: &hostPathDirectory,
								},
							},
						},
					},
				},
			},
		},
	}
}
