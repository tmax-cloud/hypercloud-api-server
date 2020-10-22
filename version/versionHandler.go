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
	podList, exist := k8sApiCaller.GetPodListByLabel("hypercloud4=operator")
	if exist {
		mod.HyperCloudOperator.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.HyperCloudOperator.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.HyperCloudOperator.Status = "not found"
	}

	// hypercloud console version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=console")
	if exist {
		mod.HyperCloudConsole.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.HyperCloudConsole.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.HyperCloudConsole.Status = "not found"
	}

	// hypercloud webhook version
	podList, exist = k8sApiCaller.GetPodListByLabel("hypercloud4=webhook")
	if exist {
		mod.HyperCloudWebHook.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.HyperCloudWebHook.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.HyperCloudWebHook.Status = "not found"
	}

	// hypercloud watcher version
	podList, exist = k8sApiCaller.GetPodListByLabel("hypercloud4=secret-watcher")
	if exist {
		mod.HyperCloudWatcher.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.HyperCloudWatcher.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.HyperCloudWatcher.Status = "not found"
	}

	// k8s verison
	podList, exist = k8sApiCaller.GetPodListByLabel("component=kube-apiserver")
	if exist {
		mod.Kubernetes.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.Kubernetes.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.Kubernetes.Status = "not found"
	}

	// Calico version
	podList, exist = k8sApiCaller.GetPodListByLabel("k8s-app=calico-kube-controllers")
	if exist {
		mod.Calico.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.Calico.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.Calico.Status = "not found"
	}

	// Rook-Ceph version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=rook-ceph-operator")
	if exist {
		mod.RookCeph.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.RookCeph.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.RookCeph.Status = "not found"
	}

	// Prometheus version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=prometheus")
	if exist {
		mod.Prometheus.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.Prometheus.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.Prometheus.Status = "not found"
	}

	// Grafana version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=grafana")
	if exist {
		mod.Grafana.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.Grafana.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.Grafana.Status = "not found"
	}

	// Tekton version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=tekton-pipelines-controller")
	if exist {
		mod.Tekton.Version = podList.Items[0].Labels["version"]
		mod.Tekton.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.Tekton.Status = "not found"
	}

	// Tekton Trigger version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=tekton-triggers-controller")
	if exist {
		mod.TektonTrigger.Version = podList.Items[0].Labels["version"]
		mod.TektonTrigger.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.TektonTrigger.Status = "not found"
	}

	// Catalog controller version
	podList, exist = k8sApiCaller.GetPodListByLabel("app=catalog-catalog-controller-manager")
	if exist {
		mod.CatalogController.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.CatalogController.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.CatalogController.Status = "not found"
	}

	// TemplateServiceBroker version
	podList, exist = k8sApiCaller.GetPodListByLabel("hypercloud4=template-service-broker")
	if exist {
		mod.TemplateServiceBroker.Version = ParsingVersion(podList.Items[0].Spec.Containers[0].Image)
		mod.TemplateServiceBroker.Status = string(podList.Items[0].Status.Phase)
	} else {
		mod.TemplateServiceBroker.Status = "not found"
	}

	// encode to JSON format and response
	util.SetResponse(res, "", mod, http.StatusOK)
}

func ParsingVersion(str string) string {
	slice := strings.Split(str, ":")
	return slice[len(slice)-1]
}
