package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	alert "github.com/tmax-cloud/hypercloud-api-server/alert"
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
	mux.HandleFunc("/user", serveUser)
	mux.HandleFunc("/metering", serveMetering)
	mux.HandleFunc("/namespace", serveNamespace)
	mux.HandleFunc("/alert", serveAlert)
	mux.HandleFunc("/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/version", serveVersion)

	// HTTP Server Start
	klog.Info("Starting Hypercloud-Operator-API server...")
	klog.Flush()

	if err := http.ListenAndServe(":80", mux); err != nil {
		klog.Errorf("Failed to listen and serve Hypercloud-Operator-API server: %s", err)
	}
	klog.Info("Started Hypercloud-Operator-API server")

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
