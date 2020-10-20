package version

import (
	"hypercloud-api-server/util"
	k8sApiCaller "hypercloud-api-server/util/Caller"
	"net/http"
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

	// hypercloud operator version

	// hypercloud console version

	// hypercloud webhook version

	// hypercloud watcher version

	// k8s verison
	k8sVersion, err := k8sApiCaller.GetK8sVersion()
	if err != nil {
		klog.Errorln("k8s version status error")
	} else {
		mod.Kubernetes.Status = "true"
	}
	mod.Kubernetes.Version = k8sVersion

	// Calico version
	podList, _ := k8sApiCaller.GetPodListByLabel("k8s-app=calico-kube-controllers")
	mod.Calico.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
	mod.Calico.Status = string(podList.Items[0].Status.Phase)

	// Rook-Ceph version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=rook-ceph-operator")
	mod.RookCeph.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
	mod.RookCeph.Status = string(podList.Items[0].Status.Phase)

	// Prometheus version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=prometheus")
	mod.Prometheus.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
	mod.Prometheus.Status = string(podList.Items[0].Status.Phase)

	// Grafana version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=grafana")
	mod.Grafana.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
	mod.Grafana.Status = string(podList.Items[0].Status.Phase)

	// Tekton version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=tekton-pipelines-controller")
	mod.Tekton.Version = podList.Items[0].Labels["version"]
	mod.Tekton.Status = string(podList.Items[0].Status.Phase)

	// Tekton Trigger version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=tekton-triggers-controller")
	mod.TektonTrigger.Version = podList.Items[0].Labels["version"]
	mod.TektonTrigger.Status = string(podList.Items[0].Status.Phase)

	// Catalog controller version
	podList, _ = k8sApiCaller.GetPodListByLabel("app=catalog-catalog-controller-manager")
	mod.CatalogController.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
	mod.CatalogController.Status = string(podList.Items[0].Status.Phase)

	// TemplateServiceBroker version

	// encode to JSON format and response
	util.SetResponse(res, "", mod, http.StatusOK)
}

func ParsingVersion(str string) string {
	slice := strings.Split(str, ":")
	return slice[len(slice)-1]
}
