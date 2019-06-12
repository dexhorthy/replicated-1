module github.com/replicatedhq/replicated

go 1.14

require (
	github.com/fatih/color v1.7.0
	github.com/go-git/go-git/v5 v5.2.0
	github.com/go-kit/kit v0.10.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/manifoldco/promptui v0.8.0
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12
	github.com/mholt/archiver/v3 v3.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/pact-foundation/pact-go v1.0.4
	github.com/pkg/errors v0.9.1
	github.com/replicatedhq/kots v1.35.0
	github.com/replicatedhq/libyaml v0.0.0-20210105211344-8240d6437a94
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tj/go-spin v1.1.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.20.5 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect

)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20170817175659-5f6282db7d65
	github.com/docker/docker => github.com/docker/docker v0.0.0-20180522102801-da99009bbb11
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/apiserver => k8s.io/apiserver v0.20.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.2
	k8s.io/code-generator => k8s.io/code-generator v0.20.2
	k8s.io/component-base => k8s.io/component-base v0.20.2
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.2
	k8s.io/cri-api => k8s.io/cri-api v0.20.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.2
	k8s.io/kubectl => k8s.io/kubectl v0.20.2
	k8s.io/kubelet => k8s.io/kubelet v0.20.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.2
	k8s.io/metrics => k8s.io/metrics v0.20.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.2
	k8s.io/node-api => k8s.io/node-api v0.20.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.2
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.2
	k8s.io/sample-controller => k8s.io/sample-controller v0.20.2
)
