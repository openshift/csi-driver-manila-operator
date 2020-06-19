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

func (r *ReconcileManilaDriver) handleNFSNodePluginDaemonSet(instance *maniladriverv1alpha1.ManilaDriver, reqLogger logr.Logger) error {
	reqLogger.Info("Reconciling NFS Node Plugin DaemonSet")

	// Define a new DaemonSet object
	ds := generateNFSNodePluginManifest()

	if err := annotator.SetLastAppliedAnnotation(ds); err != nil {
		return err
	}

	// Check if this DaemonSet already exists
	found := &appsv1.DaemonSet{}
	err := r.apiReader.Get(context.TODO(), types.NamespacedName{Name: ds.Name, Namespace: ds.Namespace}, found)
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

	// Check if we need to update the object
	equal, err := compareLastAppliedAnnotations(found, ds)
	if err != nil {
		return err
	}

	if !equal {
		reqLogger.Info("Updating DaemonSet with new changes", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
		err = r.client.Update(context.TODO(), ds)
		if err != nil {
			return err
		}
	} else {
		// DaemonSet already exists - don't requeue
		reqLogger.Info("Skip reconcile: DaemonSet already exists", "DaemonSet.Namespace", found.Namespace, "DaemonSet.Name", found.Name)
	}

	return nil
}

func generateNFSNodePluginManifest() *appsv1.DaemonSet {
	trueVar := true

	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	hostPathDirectory := corev1.HostPathDirectory

	mountPropagationBidirectional := corev1.MountPropagationBidirectional

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "csi-nodeplugin-nfsplugin",
			Namespace: "openshift-manila-csi-driver",
			Labels:    labelsNFSNodePlugin,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsNFSNodePlugin,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsNFSNodePlugin,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "csi-nodeplugin",
					Containers: []corev1.Container{
						{
							Name:  "nfs",
							Image: getCSIDriverNFSImage(),
							Args: []string{
								"--nodeid=$(NODE_ID)",
								"--endpoint=unix://plugin/csi.sock",
								"--mount-permissions=0777",
							},
							Env: []corev1.EnvVar{
								{
									Name: "NODE_ID",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged:               &trueVar,
								AllowPrivilegeEscalation: &trueVar,
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"SYS_ADMIN",
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "plugin-dir",
									MountPath: "/plugin",
								},
								{
									Name:             "pods-mount-dir",
									MountPath:        "/var/lib/kubelet/pods",
									MountPropagation: &mountPropagationBidirectional,
								},
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "plugin-dir",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kubelet/plugins/csi-nfsplugin",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "pods-mount-dir",
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
