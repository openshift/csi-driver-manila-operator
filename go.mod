module github.com/openshift/csi-driver-manila-operator

go 1.13

require (
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gophercloud/gophercloud v0.6.1-0.20191122030953-d8ac278c1c9d
	github.com/gophercloud/utils v0.0.0-20200508015959-b0167b94122c
	github.com/openshift/api v0.0.0-20200723165920-787191c0c3c8
	github.com/openshift/build-machinery-go v0.0.0-20200713135615-1f43d26dccc7
	github.com/openshift/library-go v0.0.0-20200723181926-829d4689a31f
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.0-rc.2
	k8s.io/apimachinery v0.19.0-rc.2
	k8s.io/client-go v0.19.0-rc.2
	k8s.io/code-generator v0.19.0-rc.2
	k8s.io/component-base v0.19.0-rc.2
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.2.0
)
