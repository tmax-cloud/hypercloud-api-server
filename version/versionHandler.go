package version

import (
	"hypercloud-api-server/util"
	k8sApiCaller "hypercloud-api-server/util/Caller"
	"net/http"
	"reflect"
	"strings"

	"k8s.io/klog"
)

type Version struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type Module struct {
	HyperCloudOperator    Version `json:"hypercloudoperator"`
	HyperCloudConsole     Version `json:"hypercloudconsole"`
	HyperCloudWebHook     Version `json:"hypercloudwebhook"`
	HyperCloudWatcher     Version `json:"hypercloudwatcher"`
	Kubernetes            Version `json:"kubernetes"`
	Calico                Version `json:"calico"`
	RookCeph              Version `json:"rookceph"`
	Prometheus            Version `json:"prometheus"`
	Grafana               Version `json:"grafana"`
	Tekton                Version `json:"tekton"`
	TektonTrigger         Version `json:"tektontrigger"`
	CatalogController     Version `json:"catalogcontroller"`
	TemplateServiceBroker Version `json:"templateservicebroker"`
}

func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET /version")

	var mod Module

	// appName indicates the label of specific pod.
	// If the form of label is changed, fix this string array.
	// The order must be matched with moudule struct.
	appName := []string{
		"hypercloud4=operator",                   // HyperCloud Operator
		"app=console",                            // HyperCloud Console
		"hypercloud4=webhook",                    // HyperCloud WebHook
		"hypercloud4=secret-watcher",             // HyperCloud Watcher
		"component=kube-apiserver",               // Kubernetes
		"k8s-app=calico-kube-controllers",        // Calico
		"app=rook-ceph-operator",                 // Rook-Ceph
		"app=prometheus",                         // Prometheus
		"app=grafana",                            // Grafana
		"app=tekton-pipelines-controller",        // Tekton
		"app=tekton-triggers-controller",         // Tekton Trigger
		"app=catalog-catalog-controller-manager", // Catalog Controller
		"hypercloud4=template-service-broker",    // TemplateServiceBroker
	}

	values := reflect.ValueOf(&mod)
	num := reflect.TypeOf(mod).NumField()

	// get version according to the order of 'Module' struct
	for i := 0; i < num; i++ {
		podList, exist := k8sApiCaller.GetPodListByLabel(appName[i])
		module := reflect.Indirect(values).Field(i) // specific module in 'mod' structure

		if exist {
			module.FieldByName("Status").SetString(string(podList.Items[0].Status.Phase))
			if podList.Items[0].Labels["version"] != "" {
				// If there is 'version' key value in label, take it.
				module.FieldByName("Version").SetString(podList.Items[0].Labels["version"])
			} else {
				// If not, take version information from image tag
				module.FieldByName("Version").SetString(ParsingVersion(podList.Items[0].Spec.Containers[0].Image))
			}
		} else {
			module.FieldByName("Status").SetString("not found")
		}
	}

	// encode to JSON format and response
	util.SetResponse(res, "", mod, http.StatusOK)
}

func ParsingVersion(str string) string {
	slice := strings.Split(str, ":")
	return slice[len(slice)-1]
}
