module github.com/openshift/csi-driver-manila-operator

go 1.15

require (
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gophercloud/gophercloud v0.15.0
	github.com/gophercloud/utils v0.0.0-20210113034859-6f548432055a
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/openshift/build-machinery-go v0.0.0-20200917070002-f171684f77ab
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/openshift/library-go v0.0.0-20210113192829-cfbb3f4c80c2
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.0
	k8s.io/component-base v0.20.0
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/yaml v1.2.0
)
