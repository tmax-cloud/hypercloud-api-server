module github.com/tmax-cloud/hypercloud-api-server

go 1.15

require (
	github.com/Shopify/sarama v1.19.0
	github.com/aws/aws-sdk-go v1.27.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/uuid v1.1.1
	// github.com/gorilla/mux v1.7.3
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.9.0
	github.com/oklog/ulid v1.3.1
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/robfig/cron v1.2.0
	github.com/tmax-cloud/efk-operator v0.0.0-20201207030412-fd9c02a3e1c2
	github.com/tmax-cloud/hypercloud-multi-operator v0.0.0-20210225041531-7124d026aacf
	github.com/tmax-cloud/hypercloud-single-operator v0.0.0-20210222045913-0ace319d7c34
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/apiserver v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.19.3
	k8s.io/utils v0.0.0-20201005171033-6301aaf42dc7
	sigs.k8s.io/controller-runtime v0.7.0
)
