package maniladriver

import (
	"context"

	"github.com/go-logr/logr"
	maniladriverv1alpha1 "github.com/openshift/csi-driver-manila-operator/pkg/apis/maniladriver/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileManilaDriver) handleManilaControllerPluginDeployment(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling Manila Controller Plugin Deployment")

	// Define a new Deployment object
	ss := generateManilaControllerPluginDeployment()

	if err := annotator.SetLastAppliedAnnotation(ss); err != nil {
		return err
	}

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", ss.Namespace, "Deployment.Name", ss.Name)
		err = r.client.Create(context.TODO(), ss)
		if err != nil {
			return err
		}

		// Deployment created successfully - don't requeue
		return nil
	} else if err != nil {
		return err
	}

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, ss)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating Deployment with new changes", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		err = r.client.Update(context.TODO(), ss)
		if err != nil {
			return err
		}
	} else {
		// Deployment already exists - don't requeue
		reqLogger.Info("Skip reconcile: Deployment already exists", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	}

	return nil
}

func generateManilaControllerPluginDeployment() *appsv1.Deployment {
	trueVar := true
	replicaNumber := int32(1)
	mountPropagationBidirectional := corev1.MountPropagationBidirectional
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	hostPathDirectory := corev1.HostPathDirectory

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openstack-manila-csi-controllerplugin",
			Namespace: "openshift-manila-csi-driver",
			Labels:    labelsManilaControllerPlugin,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaNumber,
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
