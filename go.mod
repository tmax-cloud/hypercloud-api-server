module hypercloud-api-server

go 1.15

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.1
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/robfig/cron v1.2.0
	github.com/tmax-cloud/hypercloud-go-operator v0.0.0-20201125074013-0e686fd12999
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/apiserver v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.19.2
	sigs.k8s.io/structured-merge-diff v1.0.2
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)
