module github.com/openshift/csi-driver-manila-operator

go 1.13

require (
	github.com/banzaicloud/k8s-objectmatcher v1.3.2
	github.com/go-logr/logr v0.1.0
	github.com/gophercloud/gophercloud v0.6.1-0.20191122030953-d8ac278c1c9d
	github.com/gophercloud/utils v0.0.0-20200324021909-95fb81d3291f
	github.com/nsf/jsondiff v0.0.0-20200515183724-f29ed568f4ce
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/cloud-credential-operator v0.0.0-20200406220359-beb5844a1e05
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.17.4 // Required by prometheus-operator
)
