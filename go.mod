module virt-webui

go 1.15

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628

require (
	github.com/astaxie/beego v1.12.2
	github.com/spf13/pflag v1.0.3
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	kubevirt.io/client-go v0.19.0
)
