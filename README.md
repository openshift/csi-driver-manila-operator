# CSI Driver Manila Operator

Operator to create, configure and manage CSI driver for OpenStack Manila in OpenShift.

## Quick Start

### Installing the operator

The operator needs its own namespace, service account, security context, and a few roles and bindings. For example, to install these on OpenShift >= 4.4:

```sh
oc apply -f deploy/namespace.yaml -f deploy/crds/csi.openshift.io_maniladrivers_crd.yaml -f deploy/service_account.yaml -f deploy/role_binding.yaml -f deploy/role.yaml -f deploy/operator.yaml
```

You can check logs of the operator by executing:

```sh
oc logs -f -n openshift-manila-csi-driver-operator $(oc get pods --no-headers -n openshift-manila-csi-driver-operator -o custom-columns=":metadata.name")
```

### Installing the driver

When the operator is started, you need to create a CR to install the driver:

```sh
oc apply -f deploy/crds/csi.openshift.io_v1alpha1_maniladriver_cr.yaml
```

**Note:** ManilaDriver CR is a singleton, which means you can't create more than one instance of this resource. By convention this should be a cluster-scoped object called `cluster`.

Operator automatically creates required StorageClasses for all Manila share types. Each of them is called `manila-csi-<share_type>`.

To see the list of provisioned Storage Classes execute:

```sh
oc get storageclasses
```

All driver's resources are created in  the `openshift-manila-csi-driver` namespace.

### Creating PVCs and Pods

You're all set now! However, you likely want to test the deployment, so let's create a PVC and POD for testing.

**Note:** In the PVC example we use `manila-csi-default` Storage Class, which may be different in your case.

```sh
oc create namespace manila-test
oc create -n manila-test -f examples/nfs/dynamic-provisioning/pvc.yaml
```

At this moment Manila CSI driver should provision a volume in the Manila service.

Next step is to create a pod.

```sh
oc create -n manila-test -f examples/nfs/dynamic-provisioning/pod.yaml
```

Once the pvc and pod are up and running, it will look like this:

```sh
$ oc get pod new-nfs-share-pod
NAME                READY   STATUS    RESTARTS   AGE
new-nfs-share-pod   1/1     Running   0          106s

$ oc get pvc -n manila-test
NAME                STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS     AGE
new-nfs-share-pvc   Bound    pvc-b1e5ebb8-8032-4722-92e3-06bd7ce5afec   1Gi        RWX            csi-manila-nfs   118s

$ oc get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                       STORAGECLASS     REASON   AGE
pvc-b1e5ebb8-8032-4722-92e3-06bd7ce5afec   1Gi        RWX            Delete           Bound    manila-test/new-nfs-share-pvc   csi-manila-nfs            2m50s

$ oc describe pod new-nfs-share-pod -n manila-test | grep Volumes: -A 4
Volumes:
  mypvc:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  new-nfs-share-pvc
    ReadOnly:   false
```

Looking inside the container you will notice that the provided volume has been mounted:

```sh
$ oc exec -n manila-test -it new-nfs-share-pod  -- mount | grep /var/lib/www
10.0.128.27:/volumes/_nogroup/e3c5f7fd-aeee-4485-9a40-6a732d55f689 on /var/lib/www type nfs4 (rw,relatime,vers=4.1,rsize=1048576,wsize=1048576,namlen=255,hard,proto=tcp,timeo=600,retrans=2,sec=sys,clientaddr=10.129.2.9,local_lock=none,addr=10.0.128.27)
```

### Delete the testing pod and pvc

Eventually you want to remove all the testing resources from your cluster. To do so just delete the namespace:

```sh
oc delete namespace manila-test
```

Manila provisioner will automatically delete the share in Manila service as well.

### Removing the driver and operator

First, remove the CR. The driver and its cluster-scoped resources will be deleted along with it.

```sh
oc delete -f deploy/crds/csi.openshift.io_v1alpha1_maniladriver_cr.yaml
```

When the driver is deleted, remove the remaining parts of the operator.

```sh
oc delete -f deploy/crds/csi.openshift.io_maniladrivers_crd.yaml -f deploy/role.yaml -f deploy/role_binding.yaml -f deploy/service_account.yaml -f deploy/namespace.yaml
```
