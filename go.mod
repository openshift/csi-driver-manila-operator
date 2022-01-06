module github.com/openshift/csi-driver-manila-operator

go 1.16

require (
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gophercloud/gophercloud v0.6.1-0.20191122030953-d8ac278c1c9d
	github.com/gophercloud/utils v0.0.0-20200508015959-b0167b94122c
	github.com/openshift/api v0.0.0-20211215120111-7c47a5f63470
	github.com/openshift/build-machinery-go v0.0.0-20211213093930-7e33a7eb4ce3
	github.com/openshift/client-go v0.0.0-20211209144617-7385dd6338e3
	github.com/openshift/library-go v0.0.0-20211222155012-624c91f4e514
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v0.23.1
	k8s.io/component-base v0.23.1
	k8s.io/klog/v2 v2.30.0
	sigs.k8s.io/yaml v1.2.0
)
