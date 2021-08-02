package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	gmux "github.com/gorilla/mux"
	admission "github.com/tmax-cloud/hypercloud-api-server/admission"
	"github.com/tmax-cloud/hypercloud-api-server/alert"
	audit "github.com/tmax-cloud/hypercloud-api-server/audit"
	awscost "github.com/tmax-cloud/hypercloud-api-server/awscost"
	cluster "github.com/tmax-cloud/hypercloud-api-server/cluster"
	claim "github.com/tmax-cloud/hypercloud-api-server/clusterClaim"
	metering "github.com/tmax-cloud/hypercloud-api-server/metering"
	"github.com/tmax-cloud/hypercloud-api-server/namespace"
	"github.com/tmax-cloud/hypercloud-api-server/namespaceClaim"
	user "github.com/tmax-cloud/hypercloud-api-server/user"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	"github.com/tmax-cloud/hypercloud-api-server/util/caller"
	kafkaConsumer "github.com/tmax-cloud/hypercloud-api-server/util/consumer"
	version "github.com/tmax-cloud/hypercloud-api-server/version"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"

	"net/http"

	"github.com/robfig/cron"
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
	flag.StringVar(&util.SMTPHost, "smtpHost", "mail.tmax.co.kr", "SMTP Server Host Address")
	flag.IntVar(&util.SMTPPort, "smtpPort", 25, "SMTP Server Port")
	flag.StringVar(&util.SMTPUsernamePath, "smtpUsername", "/run/secrets/smtp/username", "SMTP Server Username")
	flag.StringVar(&util.SMTPPasswordPath, "smtpPassword", "/run/secrets/smtp/password", "SMTP Server Password")
	flag.StringVar(&util.AccessSecretPath, "accessSecret", "/run/secrets/token/accessSecret", "Token Access Secret")
	flag.StringVar(&util.HtmlHomePath, "htmlPath", "/run/configs/html/", "Invite htlm path")
	// flag.StringVar(&util.TokenExpiredDate, "tokenExpiredDate", "24hours", "Token Expired Date")

	// Get Hypercloud Operating Mode!!!
	hcMode := os.Getenv("HC_MODE")
	util.TokenExpiredDate = os.Getenv("INVITATION_TOKEN_EXPIRED_DATE")
	kafkaConsumer.KafkaGroupId = os.Getenv("KAFKA_GROUP_ID")
	util.ReadFile()
	caller.UpdateAuditResourceList()

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
		err = ioutil.WriteFile("./logs/api-server"+time.Now().Format("2006-01-02")+".log", input, 0644)
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
	// cronJob.AddFunc("@hourly", audit.UpdateAuditResource)
	cronJob.Start()

	// Hyperauth Event Consumer
	go kafkaConsumer.HyperauthConsumer()

	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		klog.Errorf("Failed to load key pair: %s", err)
	}

	// Req multiplexer
	mux := gmux.NewRouter()
	// mux := http.NewServeMux()
	mux.HandleFunc("/user", serveUser)
	mux.HandleFunc("/metering", serveMetering)
	mux.HandleFunc("/namespace", serveNamespace)
	mux.HandleFunc("/alert", serveAlert)
	mux.HandleFunc("/grafanaUser", serveGrafanaUser)
	mux.HandleFunc("/grafanaDashboard", serveGrafanaDashboard)
	mux.HandleFunc("/namespaceClaim", serveNamespaceClaim)
	mux.HandleFunc("/version", serveVersion)
	mux.HandleFunc("/awscost", serveAwscost)

	if hcMode != "single" {
		// for multi mode only
		mux.HandleFunc("/clusterclaims", serveClusterClaim)                                       // List all clusterclaim
		mux.HandleFunc("/namespaces/{namespace}/clusterclaims", serveClusterClaim)                // list all clusterclaim in a specific namespace
		mux.HandleFunc("/namespaces/{namespace}/clusterclaims/{clusterclaim}", serveClusterClaim) // Admit clusterclaim request
		mux.HandleFunc("/clustermanagers", serveCluster)
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers", serveCluster) // list clustermanager for all namespaces (list page & all ns)
		mux.HandleFunc("/clustermanagers/{access}", serveCluster)               // list accessible clustermanager for all namespaces (lnb & all ns)
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers", serveCluster) // list all clustermanager in a specific namespace
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}", serveCluster)
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/member", serveClusterMember)                                     // list all member
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/member_invitation", serveClusterInvitation)                      // list a pending status user
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/member_invitation/{attribute}/{member}", serveClusterInvitation) // 추가 요청 (db + token 발급)
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/member_invitation/{admit}", serveClusterInvitationAdmit)         // 추가 요청 승인, 추가 요청 거절
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/remove_member/{attribute}/{member}", serveClusterMember)         // 멤버 삭제 (db)
		mux.HandleFunc("/namespaces/{namespace}/clustermanagers/{clustermanager}/update_role/{attribute}/{member}", serveClusterMember)           // 권한 변경 (db)
	}

	mux.HandleFunc("/metadata", serveMetadata)
	mux.HandleFunc("/audit/member_suggestions", serveAuditMemberSuggestions)
	mux.HandleFunc("/audit", serveAudit)
	mux.HandleFunc("/audit/{clusterNamespace}/{clusterName}", serveAudit)
	mux.HandleFunc("/audit/batch", serveAuditBatch)
	mux.HandleFunc("/audit/resources", serveAuditResources)
	mux.HandleFunc("/audit/cluster", serveAuditCluster)
	mux.HandleFunc("/audit/verb", serveAuditVerb)
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
	klog.Info("Setiing Grafana Admin...")
	hc_admin := caller.GetCRBAdmin()
	caller.CreateGrafanaUser(hc_admin)
	id := caller.GetGrafanaUser(hc_admin)
	adminBody := `{"isGrafanaAdmin": true}`
	grafanaId, grafanaPw := "admin", "admin"
	httpgeturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/admin/users/" + strconv.Itoa(id) + "/permissions"

	request, _ := http.NewRequest("PUT", httpgeturl, bytes.NewBuffer([]byte(adminBody)))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		resbody, _ := ioutil.ReadAll(response.Body)
		klog.Infof(string(resbody))
	}

	//get grafana key
	httpposturl := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/auth/keys"
	var GrafanaKeyBody util.GrafanaKeyBody

	GrafanaKeyBody.Name = caller.RandomString(8)
	GrafanaKeyBody.Role = "Admin"
	GrafanaKeyBody.SecondsToLive = 300
	json_body, _ := json.Marshal(GrafanaKeyBody)
	request, _ = http.NewRequest("POST", httpposturl, bytes.NewBuffer(json_body))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		body, _ := ioutil.ReadAll(response.Body)

		klog.Infof(string(body))
		var grafana_resp util.Grafana_key
		json.Unmarshal([]byte(body), &grafana_resp)
		util.GrafanaKey = "Bearer " + grafana_resp.Key
		klog.Infof(util.GrafanaKey)
	}

	//org permission
	httpgeturlorg := "http://" + grafanaId + ":" + grafanaPw + "@" + util.GRAFANA_URI + "api/orgs/1/users/" + strconv.Itoa(id)
	adminorgBody := `{"role":"Admin"}`
	request, _ = http.NewRequest("PATCH", httpgeturlorg, bytes.NewBuffer([]byte(adminorgBody)))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	//request.Header.Set("Authorization", util.GrafanaKey)
	client2 := &http.Client{}
	response, err = client2.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		resbody, _ := ioutil.ReadAll(response.Body)

		klog.Infof(string(resbody))
	}

	//default dashboard permission to only admin

	klog.Infof("default dashboard permission setting(admin)")
	permBody := `{
		"items": []
	}`
	httpposturl_per := "http://" + util.GRAFANA_URI + "api/dashboards/id/1/permissions"
	request, _ = http.NewRequest("POST", httpposturl_per, bytes.NewBuffer([]byte(permBody)))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", util.GrafanaKey)
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		resbody, _ := ioutil.ReadAll(response.Body)
		klog.Infof(string(resbody))
	}
	httpposturl_per = "http://" + util.GRAFANA_URI + "api/dashboards/id/2/permissions"
	request, _ = http.NewRequest("POST", httpposturl_per, bytes.NewBuffer([]byte(permBody)))

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Authorization", util.GrafanaKey)
	client = &http.Client{}
	response, err = client.Do(request)
	if err != nil {
		klog.Errorln(err)

	} else {
		defer response.Body.Close()
		resbody, _ := ioutil.ReadAll(response.Body)
		klog.Infof(string(resbody))
	}
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
func serveGrafanaUser(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodPost:
		caller.CreateGrafanaUser("test12@tmax.co.kr")
	default:
		//error
	}
}

func serveGrafanaDashboard(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodPost:
		caller.CreateDashBoard(res, req)
	case http.MethodDelete:
		caller.DeleteGrafanaDashboard(res, req)
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
	vars := gmux.Vars(req)
	switch req.Method {
	case http.MethodGet:
		if vars["access"] == "access" {
			cluster.ListLNB(res, req)
		} else if vars["access"] == "" {
			cluster.ListPage(res, req)
		} else {
			// errror
		}
	case http.MethodPost:
		cluster.InsertCLM(res, req)
	case http.MethodDelete:
		cluster.DeleteCLM(res, req)
	default:
	}
}

func serveClusterMember(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	switch req.Method {
	case http.MethodPost:
		cluster.RemoveMember(res, req)
	case http.MethodGet:
		cluster.ListClusterMember(res, req)
	case http.MethodPut:
		cluster.UpdateMemberRole(res, req)
	default:
	}
}

func serveClusterInvitation(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	vars := gmux.Vars(req)
	switch req.Method {
	case http.MethodGet:
		cluster.ListInvitation(res, req)
		break
	case http.MethodPost:
		if vars["attribute"] == "user" {
			cluster.InviteUser(res, req)
		} else if vars["attribute"] == "group" {
			cluster.InviteGroup(res, req)
		} else {
			// errror
		}
		break
	default:
	}
}

func serveClusterInvitationAdmit(res http.ResponseWriter, req *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", req.Method, req.URL.Path)
	vars := gmux.Vars(req)
	switch req.Method {
	case http.MethodGet:
		if vars["admit"] == "accept" {
			cluster.AcceptInvitation(res, req)
		} else if vars["admit"] == "reject" {
			cluster.DeclineInvitation(res, req)
		} else {
			// errror
		}
		break
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

func serveAuditVerb(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		audit.ListAuditVerb(w, r)
	default:
		//error
	}
}

func serveAuditResources(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		audit.ListAuditResource(w, r)
	default:
		//error
	}
}

func serveAuditCluster(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		audit.ListAuditCluster(w, r)
	default:
		//error
	}
}

func serveAuditMemberSuggestions(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Http request: method=%s, uri=%s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		audit.MemberSuggestions(w, r)
	default:
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
	// klog.Infof("Request body: %s\n", body)

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

	// klog.Infof("Response body: %s\n", respBytes)

	if err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = admission.ToAdmissionResponse(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = admission.ToAdmissionResponse(err)
	}
}

func serveAwscost(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		awscost.Get(res, req)
	case http.MethodPut:
	case http.MethodOptions:
	default:
		//error
	}
}
