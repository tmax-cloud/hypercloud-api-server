package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	alert "github.com/tmax-cloud/hypercloud-api-server/alert"
	cluster "github.com/tmax-cloud/hypercloud-api-server/cluster"
	claim "github.com/tmax-cloud/hypercloud-api-server/clusterClaim"
	metering "github.com/tmax-cloud/hypercloud-api-server/metering"
	"github.com/tmax-cloud/hypercloud-api-server/namespace"
	"github.com/tmax-cloud/hypercloud-api-server/namespaceClaim"
	user "github.com/tmax-cloud/hypercloud-api-server/user"
	version "github.com/tmax-cloud/hypercloud-api-server/version"

	"k8s.io/klog"

	"net/http"

	"github.com/robfig/cron"
)

func main() {
	// Get Hypercloud Operating Mode!!!
	hcMode := os.Getenv("HC_MODE")

	// For Log file
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	file, err := os.OpenFile(
		"./logs/api-server.log",
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		os.FileMode(0644),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	w := io.MultiWriter(file, os.Stdout)
	klog.SetOutput(w)

	// Logging Cron Job
	cronJob := cron.New()
	cronJob.AddFunc("1 0 0 * * ?", func() {
		input, err := ioutil.ReadFile("./logs/api-server.log")
		if err != nil {
			klog.Error(err)
			return
		}
		err = ioutil.WriteFile("./logs/api-server"+time.Now().AddDate(0, 0, -1).Format("2006-01-02")+".log", input, 0644)
		if err != nil {
			klog.Error(err, "Error creating", "./logs/api-server")
			fmt.Println(err)
			return
		}
		klog.Info("Log BackUp Success")
		os.Truncate("./logs/api-server.log", 0)
		file.Seek(0, os.SEEK_SET)
	})

	// Metering Cron Job
	cronJob.AddFunc("0 */1 * ? * *", metering.MeteringJob)
	cronJob.Start()

	// Req multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hypercloud/user", serveUser)
	mux.HandleFunc("/api/hypercloud/metering", serveMetering)
	mux.HandleFunc("/api/hypercloud/namespace", serveNamespace)
	mux.HandleFunc("/api/hypercloud/alert", serveAlert)
	mux.HandleFunc("/api/hypercloud/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/api/hypercloud/version", serveVersion)

	if hcMode != "single" {
		// for multi mode only
		mux.HandleFunc("/api/hypercloud/clusterclaim", serveClusterClaim)
		mux.HandleFunc("/api/hypercloud/cluster", serveCluster)
		mux.HandleFunc("/api/hypercloud/cluster/owner", serveClusterOwner)
		mux.HandleFunc("/api/hypercloud/cluster/member", serveClusterMember)
		mux.HandleFunc("/api/hypercloud/test/", serveTest)
	}

	// HTTP Server Start
	klog.Info("Starting Hypercloud5-API server...")
	klog.Flush()

	if err := http.ListenAndServe(":80", mux); err != nil {
		klog.Errorf("Failed to listen and serve Hypercloud5-API server: %s", err)
	}
	klog.Info("Started Hypercloud5-API server")

}

func serveNamespace(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespace.Get(res, req)
	case http.MethodPut:
		namespace.Put(res, req)
	case http.MethodOptions:
		namespace.Options(res, req)
	default:
		//error
	}
}

func serveNamespaceClaim(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		namespaceClaim.Get(res, req)
	case http.MethodPut:
		namespaceClaim.Put(res, req)
	case http.MethodOptions:
		namespaceClaim.Options(res, req)
	default:
		//error
	}
}

func serveUser(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		user.Post(res, req)
	case http.MethodDelete:
		user.Delete(res, req)
	case http.MethodOptions:
		user.Options(res, req)
	default:
		//error
	}
}

func serveMetering(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		metering.Get(res, req)
	case http.MethodOptions:
		metering.Options(res, req)
	default:
		//error
	}
}

func serveAlert(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		alert.Post(res, req)
	case http.MethodGet:
		alert.Get(res, req)
	default:
		//error
	}
}
func serveVersion(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		version.Get(res, req)
	default:
		//error
	}
}

func serveTest(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	klog.Info("Request body: \n", string(body))
}

func serveClusterClaim(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		//curl -XPUT 172.22.6.2:32319/api/master/clusterclaim?userId=sangwon_cho@tmax.co.kr
		claim.List(res, req)
	case http.MethodPost:
	case http.MethodPut:
		//curl -XPUT 172.22.6.2:32319/api/master/clusterclaim?userId=sangwon_cho@tmax.co.kr\&clusterClaim=test-d5n92\&admit=true
		claim.Put(res, req)
	case http.MethodDelete:
	default:
	}
}

func serveCluster(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.List(res, req)
	case http.MethodPost:
	case http.MethodPut:
		// invite multiple users
		cluster.Put(res, req)
	case http.MethodDelete:
	default:
	}
}

func serveClusterOwner(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.ListOwner(res, req)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	default:
	}
}

func serveClusterMember(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.ListMember(res, req)
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodDelete:
	default:
	}
}
