module github.com/openshift/csi-driver-manila-operator

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gophercloud/gophercloud v0.6.1-0.20191122030953-d8ac278c1c9d
	github.com/gophercloud/utils v0.0.0-20200508015959-b0167b94122c
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/openshift/api v0.0.0-20210331193751-3acddb19d360
	github.com/openshift/build-machinery-go v0.0.0-20210209125900-0da259a2c359
	github.com/openshift/client-go v0.0.0-20210331195552-cf6c2669e01f
	github.com/openshift/library-go v0.0.0-20210408164723-7a65fdb398e2
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0-rc.0
	k8s.io/apimachinery v0.21.0-rc.0
	k8s.io/client-go v0.21.0-rc.0
	k8s.io/component-base v0.21.0-rc.0
	k8s.io/klog/v2 v2.8.0
	sigs.k8s.io/yaml v1.2.0
)
