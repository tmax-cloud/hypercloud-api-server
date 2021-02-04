package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	admission "github.com/tmax-cloud/hypercloud-api-server/admission"
	"github.com/tmax-cloud/hypercloud-api-server/alert"
	audit "github.com/tmax-cloud/hypercloud-api-server/audit"
	cluster "github.com/tmax-cloud/hypercloud-api-server/cluster"
	claim "github.com/tmax-cloud/hypercloud-api-server/clusterClaim"
	metering "github.com/tmax-cloud/hypercloud-api-server/metering"
	"github.com/tmax-cloud/hypercloud-api-server/namespace"
	"github.com/tmax-cloud/hypercloud-api-server/namespaceClaim"
	user "github.com/tmax-cloud/hypercloud-api-server/user"
	version "github.com/tmax-cloud/hypercloud-api-server/version"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"

	"net/http"

	"github.com/robfig/cron"
	// kafkaConsumer "github.com/tmax-cloud/hypercloud-api-server/util/Consumer"
)

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

var (
	port     int
	certFile string
	keyFile  string
)

func main() {
	// For tls
	flag.IntVar(&port, "port", 443, "hypercloud5-api-server port")
	flag.StringVar(&certFile, "certFile", "/run/secrets/tls/hypercloud-api-server.crt", "hypercloud5-api-server cert")
	flag.StringVar(&keyFile, "keyFile", "/run/secrets/tls/hypercloud-api-server.key", "hypercloud5-api-server key")
	flag.StringVar(&admission.SidecarContainerImage, "sidecarImage", "fluent/fluent-bit:1.5-debug", "Fluent-bit image name.")

	// Get Hypercloud Operating Mode!!!
	hcMode := os.Getenv("HC_MODE")

	// For Log file
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Parse()

	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", os.ModeDir)
	}

	file, err := os.OpenFile(
		"./logs/api-server.log",
		os.O_CREATE|os.O_RDWR|os.O_TRUNC,
		os.FileMode(0644),
	)
	if err != nil {
		klog.Error(err, "Error Open", "./logs/api-server")
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
			return
		}
		klog.Info("Log BackUp Success")
		os.Truncate("./logs/api-server.log", 0)
		file.Seek(0, os.SEEK_SET)
	})

	// Metering Cron Job
	cronJob.AddFunc("0 */1 * ? * *", metering.MeteringJob)
	cronJob.Start()

	// // Hyperauth Event Consumer
	// go kafkaConsumer.HyperauthConsumer()

	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %s", err)
	}

	// Req multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/user", serveUser)
	mux.HandleFunc("/metering", serveMetering)
	mux.HandleFunc("/namespace", serveNamespace)
	mux.HandleFunc("/alert", serveAlert)
	mux.HandleFunc("/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/version", serveVersion)

	if hcMode != "single" {
		// for multi mode only
		mux.HandleFunc("/clusterclaim", serveClusterClaim)
		mux.HandleFunc("/cluster", serveCluster)
		mux.HandleFunc("/cluster/remove_member", serveClusterRemoveMember)
		mux.HandleFunc("/cluster/member", serveClusterMember)
	}

	mux.HandleFunc("/metadata", serveMetadata)
	mux.HandleFunc("/audit", serveAudit)
	mux.HandleFunc("/audit/batch", serveAuditBatch)
	mux.HandleFunc("/audit/websocket", serveAuditWss)
	mux.HandleFunc("/inject/pod", serveSidecarInjectionForPod)
	mux.HandleFunc("/inject/deployment", serveSidecarInjectionForDeploy)
	mux.HandleFunc("/inject/replicaset", serveSidecarInjectionForRs)
	mux.HandleFunc("/inject/statefulset", serveSidecarInjectionForSts)
	mux.HandleFunc("/inject/daemonset", serveSidecarInjectionForDs)
	mux.HandleFunc("/inject/cronjob", serveSidecarInjectionForCj)
	mux.HandleFunc("/inject/job", serveSidecarInjectionForJob)
	mux.HandleFunc("/inject/test", serveSidecarInjectionForTest)
	mux.HandleFunc("/test", serveTest)

	// HTTP Server Start
	klog.Info("Starting Hypercloud5-API server...")
	klog.Flush()

	whsvr := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   mux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{keyPair}},
	}

	if err := whsvr.ListenAndServeTLS("", ""); err != nil {
		klog.Errorf("Failed to listen and serve Hypercloud5-API server: %s", err)
	}
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

func serveClusterClaim(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		claim.List(res, req)
	case http.MethodPut:
		claim.Put(res, req)
	default:
	}
}

func serveCluster(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodGet:
		cluster.List(res, req)
	case http.MethodPut:
		// invite multiple users
		// cluster.Put(res, req)
	default:
	}
}
func serveClusterRemoveMember(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodPost:
		cluster.RemoveMember(res, req)
	default:
	}
}

func serveClusterMember(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodPost:
		cluster.InviteMember(res, req)
	default:
	}
}

func serveMetadata(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.AddResourceMeta)
}
func serveSidecarInjectionForPod(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForPod)
}
func serveSidecarInjectionForDeploy(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForDeploy)
}
func serveSidecarInjectionForRs(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForRs)
}
func serveSidecarInjectionForSts(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForSts)
}
func serveSidecarInjectionForDs(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForDs)
}
func serveSidecarInjectionForCj(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForCj)
}
func serveSidecarInjectionForJob(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForJob)
}
func serveSidecarInjectionForTest(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	serve(w, r, admission.InjectionForTest)
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

func serveAudit(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		audit.GetAudit(w, r)
	case http.MethodPost:
		audit.AddAudit(w, r)
	case http.MethodPut:
	case http.MethodDelete:
	default:
		//error
	}
}

func serveAuditBatch(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	audit.AddAuditBatch(w, r)
}

func serveAuditWss(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	audit.ServeWss(w, r)
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	klog.Infof("Request body: %s\n", body)

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	requestedAdmissionReview := v1beta1.AdmissionReview{}
	responseAdmissionReview := v1beta1.AdmissionReview{}

	if err := json.Unmarshal(body, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = admission.ToAdmissionResponse(err)
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	respBytes, err := json.Marshal(responseAdmissionReview)

	klog.Infof("Response body: %s\n", respBytes)

	if err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = admission.ToAdmissionResponse(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = admission.ToAdmissionResponse(err)
	}
}
