module virt-webui

go 1.15

require (
	github.com/astaxie/beego v1.12.2
	github.com/spf13/pflag v1.0.5
	k8s.io/apimachinery v0.17.1-beta.0
	kubevirt.io/client-go v0.32.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v11.0.0+incompatible
	github.com/go-kit/kit => github.com/go-kit/kit v0.3.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20191219222812-2987a591a72c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20191125132246-f6563a70e19a
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.4

	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
	k8s.io/component-base => k8s.io/component-base v0.16.4
	k8s.io/cri-api => k8s.io/cri-api v0.16.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.4
	k8s.io/klog => k8s.io/klog v0.4.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.4
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.4
	k8s.io/kubectl => k8s.io/kubectl v0.16.4
	k8s.io/kubelet => k8s.io/kubelet v0.16.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.4
	k8s.io/metrics => k8s.io/metrics v0.16.4
	k8s.io/node-api => k8s.io/node-api v0.16.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.4
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.4
	k8s.io/sample-controller => k8s.io/sample-controller v0.16.4

	kubevirt.io/containerized-data-importer => kubevirt.io/containerized-data-importer v1.10.6
)
