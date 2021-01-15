package alert

import (
	"bytes"
	"context"
	"encoding/json"
	//"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	//"path/filepath"
	"strconv"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"

	alertModel "github.com/tmax-cloud/hypercloud-api-server/alert/model"

	"github.com/oklog/ulid"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/Caller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/util/homedir"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var config *restclient.Config

type alertClient struct {
	restClient rest.Interface
	ns         string
}

func init() {
	var err error
	
	// If api-server on Process, active this code.
	// var kubeconfig2 *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig2 = flag.String("kubeconfig2", filepath.Join(home, ".kube", "config"), "/root/.kube")
	// } else {
	// 	kubeconfig2 = flag.String("kubeconfig2", "", "/root/.kube")
	// }
	// flag.Parse()
	// config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig2)

	// If api-server on Pod, active this code.
	config, err = restclient.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	config.Burst = 100
	config.QPS = 100

	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	setScheme()

}

var (
	scheme     = runtime.NewScheme()
	hostclient client.Client
)

func setScheme() {
	utilruntime.Must(alertModel.AddToScheme(scheme))
}

func Post2(res http.ResponseWriter, req *http.Request) {

	klog.Infoln("**** POST /alert")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	klog.Infoln("Get Alert Boday : %s\n", string(body))
	var v alertModel.Alertaudit
	json.Unmarshal([]byte(string(body)), &v)

	for i := 0; i < len(v.Alert); i++ {
		if v.Alert[i] == '_' {
			fmt.Printf("%s\n", v.Alert)
			v.Resource = v.Alert[0:i]
			v.Status = v.Alert[i+1 : len(v.Alert)]
		}
	}
	klog.Infoln("status : %s\nresource : %s\nalert : %s\nnamespace : %s\nmessage : %s\nname : %s\n", v.Status, v.Resource, v.Alert, v.Namespace, v.Message, v.Name)

	pop := &audit.Event{}

	pop.Kind = "Event"
	pop.APIVersion = "audit.k8s.io/v1"
	pop.Stage = "ResponseComplete"
	pop.Verb = "alert"

	pop.ObjectRef = &audit.ObjectReference{
		Resource:  v.Resource,
		Namespace: v.Namespace,
		Name:      v.Name,
	}
	// pop.ObjectRef.Resource = v.Resource
	// pop.ObjectRef.Namespace = v.Namespace
	// pop.ObjectRef.Name = v.Name
	pop.ResponseStatus = &metav1.Status{
		Status:  v.Status,
		Message: v.Message,
	}
	pbytes, _ := json.Marshal(pop)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post("172.22.6.21:api/webhook/audit/batch", "application/json", buff)
	if err != nil {
		panic(err)
	}

	//var err error
	hostclient, err = client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		str := string(respBody)
		println(str)
	}
	alertBody := &alertModel.Alert{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Alert",
			APIVersion: "tmax.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
		},
		Spec: alertModel.AlertSpec{
			Name:     "test",
			Message:  "test",
			Resource: "test",
			Kind:     v.Status,
		},
	}
	//var err2 string
	err2 := hostclient.Create(context.TODO(), alertBody)
	if err2 != nil {
		klog.Errorln(err2)
		panic(err2)
	}
}
func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET /metering")
	queryParams := req.URL.Query()

	name := queryParams.Get("name")
	label := queryParams.Get("label")
	namespace := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	var resp alertModel.Alert
	resp = k8sApiCaller.GetAlert(name, namespace, label)
	util.SetResponse(res, "", resp, http.StatusOK)
	return
}

func Post(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** POST /alert")
	req.ParseForm()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Get Alert Boday : %s\n", string(body))
	var v alertModel.Alertaudit
	json.Unmarshal([]byte(string(body)), &v)

	for i := 0; i < len(v.Alert); i++ {
		if v.Alert[i] == '_' {
			fmt.Printf("%s\n", v.Alert)
			v.Resource = v.Alert[0:i]
			v.Status = v.Alert[i+1 : len(v.Alert)]
		}
	}
	klog.Infoln("status : %s\nresource : %s\nalert : %s\nnamespace : %s\nmessage : %s\nname : %s\n", v.Status, v.Resource, v.Alert, v.Namespace, v.Message, v.Name+genUlid())

	pop := &audit.Event{}

	pop.Kind = "Event"
	pop.APIVersion = "audit.k8s.io/v1"
	pop.Stage = "ResponseComplete"
	pop.Verb = "alert"

	pop.ObjectRef = &audit.ObjectReference{
		Resource:  v.Resource,
		Namespace: v.Namespace,
		Name:      v.Name,
	}
	// pop.ObjectRef.Resource = v.Resource
	// pop.ObjectRef.Namespace = v.Namespace
	// pop.ObjectRef.Name = v.Name
	pop.ResponseStatus = &metav1.Status{
		Status:  v.Status,
		Message: v.Message,
	}
	//pbytes, _ := json.Marshal(pop)
	//buff := bytes.NewBuffer(pbytes)
	//resp, err := http.Post("172.22.6.21:api/webhook/audit/batch", "application/json", buff)
	if err != nil {
		panic(err)
	}
	//defer resp.Body.Close()
	alertBody := &alertModel.Alert{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Alert",
			APIVersion: "tmax.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Name,
			Namespace: v.Namespace,
		},
		Spec: alertModel.AlertSpec{
			Name:     v.Alert,
			Message:  v.Message,
			Resource: v.Resource,
			Kind:     v.Status,
		},
	}
	k8sApiCaller.CreateAlert(*alertBody, v.Namespace)

}
func makeTimeRange(timeUnit string, startTime string, endTime string, query string) string {
	var start int64
	end := time.Now().Unix()

	if startTime != "" {
		start, _ = strconv.ParseInt(startTime, 10, 64)
	}
	if endTime != "" {
		end, _ = strconv.ParseInt(endTime, 10, 64)
	}

	switch timeUnit {
	case "hour":
		query += "select * from metering.metering_hour"
		break
	case "day":
		query += "select * from metering.metering_day"
		break
	case "month":
		query += "select * from metering.metering_month"
		break
	case "year":
		query += "select * from metering.metering_year"
		break
	}
	query += " where metering_time between '" + time.Unix(start, 0).Format("2006-01-02 15:04:05") + "' and '" + time.Unix(end, 0).Format("2006-01-02 15:04:05") + "'"
	return query
}
func genUlid() string {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}
