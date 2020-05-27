# CSI Driver Manila Operator

Operator to create, configure and manage CSI driver for OpenStack Manila in Kubernetes and OpenShift.

## Quick Start

### Installing the operator

#### Operatorhub installation

TBD

#### Manual installation

The operator needs its own namespace, service account, security context, and a few roles and bindings. For example, to install these on OpenShift >= 4.4:

```sh
oc apply -f deploy/namespace.yaml -f deploy/crds/csi.openshift.io_maniladrivers_crd.yaml -f deploy/service_account.yaml -f deploy/role_binding.yaml -f deploy/role.yaml -f deploy/operator.yaml -f deploy/crds/csi.openshift.io_v1alpha1_maniladriver_cr.yaml
```

For Kubernetes you will also need to define user credentials to access Manila and put them in a secret. An example manifest can be found in `examples/nfs/secrets.yaml`. The result should look like:

```sh
$ cat mysecrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: csi-manila-secrets
  namespace: openshift-manila-csi-driver
stringData:
  # Mandatory
  os-authURL: "http://example.com/identity"
  os-region: "RegionOne"

  # Authentication using user credentials
  os-userName: "demo"
  os-password: "secret"
  os-projectName: "demo"
  os-domainID: "default"
  os-projectDomainID: "default"
```

Create the secret with the command:

```sh
oc apply -f mysecrets.yaml
```

### Creating PVCs and Pods

You're all set now! However, you likely want to test the deployment, so let's create a pvc and pod for testing.

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
$ oc delete namespace manila-test
```

#### Removing the operator using the operator catalog

TBD

#### Removing the operator manually

First, delete the CR. The driver and its cluster-scoped resources will be removed along with it.

```sh
oc delete -f deploy/crds/csi.openshift.io_v1alpha1_maniladriver_cr.yaml
```

When the driver is deleted, remove the ramaining parts of the operator.

```sh
oc delete -f deploy/crds/csi.openshift.io_maniladrivers_crd.yaml -f deploy/role.yaml -f deploy/role_binding.yaml -f deploy/service_account.yaml -f deploy/namespace.yaml
```
