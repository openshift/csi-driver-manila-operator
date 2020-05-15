package manilacsi

import (
	"context"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	manilacsiv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/manilacsi/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *ReconcileManilaCSI) handleManilaControllerPluginStatefulSet(instance *manilacsiv1alpha1.ManilaCSI, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Controller Plugin StatefulSet")

	// Define a new StatefulSet object
	ss := generateManilaControllerPluginStatefulSet()

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

	// Check if we need to update the object
	ss.Status = found.Status
	patchResult, err := patch.DefaultPatchMaker.Calculate(found, ss)
	if err != nil {
		return err
	}

	if !patchResult.IsEmpty() {
		reqLogger.Info("Updating StatefulSet with new changes", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name, "Changes", patchResult.String())
		err = r.client.Update(context.TODO(), ss)
		if err != nil {
			return err
		}
	} else {
		// StatefulSet already exists - don't requeue
		reqLogger.Info("Skip reconcile: StatefulSet already exists", "StatefulSet.Namespace", found.Namespace, "StatefulSet.Name", found.Name)
	}

	return nil
}

func generateManilaControllerPluginStatefulSet() *appsv1.StatefulSet {
	trueVar := true
	replicaNumber := int32(1)
	mountPropagationBidirectional := corev1.MountPropagationBidirectional
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	hostPathDirectory := corev1.HostPathDirectory

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-manila-csi-controllerplugin",
			Namespace: "manila-csi",
			Labels:    labelsManilaControllerPlugin,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "openstack-manila-csi-controllerplugin",
			Replicas:    &replicaNumber,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsManilaControllerPlugin,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsManilaControllerPlugin,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "openstack-manila-csi-controllerplugin",
					Containers: []corev1.Container{
						{
							Name: "provisioner",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueVar,
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
									},
								},
								AllowPrivilegeEscalation: &trueVar,
							},
							Image: getExternalProvisionerImage(),
							Args: []string{
								"--v=5",
								"--csi-address=$(ADDRESS)",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ADDRESS",
									Value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-dir",
									MountPath: "/var/lib/kubelet/plugins/manila.csi.openstack.org",
								},
							},
						},
						{
							Name: "snapshotter",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &trueVar,
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
									},
								},
								AllowPrivilegeEscalation: &trueVar,
							},
							Image: getExternalSnaphotterImage(),
							Args: []string{
								"--v=5",
								"--csi-address=$(ADDRESS)",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ADDRESS",
									Value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-dir",
									MountPath: "/var/lib/kubelet/plugins/manila.csi.openstack.org",
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
							Image: getCSIDriverManilaImage(),
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
									Value: "unix:///var/lib/kubelet/plugins/manila.csi.openstack.org/csi-controllerplugin.sock",
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
								{
									Name:             "pod-mounts",
									MountPath:        "/var/lib/kubelet/pods",
									MountPropagation: &mountPropagationBidirectional,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
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
						{
							Name: "pod-mounts",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/pods",
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
