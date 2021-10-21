module github.com/tmax-cloud/hypercloud-api-server

go 1.15

require (
	github.com/Shopify/sarama v1.29.0
	github.com/aws/aws-sdk-go v1.27.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	// github.com/google/uuid v1.1.1
	github.com/google/uuid v1.2.0
	// github.com/gorilla/mux v1.7.3
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.10.0
	github.com/oklog/ulid v1.3.1
	github.com/robfig/cron v1.2.0
	//github.com/robfig/cron v1.2.0
	github.com/tmax-cloud/efk-operator v0.0.0-20201207030412-fd9c02a3e1c2
	github.com/tmax-cloud/hypercloud-multi-operator/v5 v5.0.2508
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
