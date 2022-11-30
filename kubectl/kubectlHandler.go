package event

import (
	"net/http"
	"os"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"github.com/tmax-cloud/hypercloud-api-server/util/caller"
	"k8s.io/klog"
)

type KubectlInfo struct {
	Image   string `json:"image"`
	Timeout string `json:"timeout"`
}

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /kubectl")
	sleepTime := os.Getenv("KUBECTL_TIMEOUT")
	if len(sleepTime) == 0 || sleepTime == "{KUBECTL_TIMEOUT}" {
		sleepTime = "21600" // 6 hours
	}
	kl := KubectlInfo{
		Image:   util.HYPERCLOUD_KUBECTL_IMAGE,
		Timeout: sleepTime,
	}
	util.SetResponse(res, "", kl, http.StatusOK)
}

func Post(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** POST /kubectl")
	queryParams := req.URL.Query()
	userName := queryParams.Get("userName")
	if err := caller.DeployKubectlPod(userName); err != nil {
		util.SetResponse(res, "", err, http.StatusBadRequest)
	} else {
		util.SetResponse(res, "Create Kubectl Pod Success", nil, http.StatusOK)
	}
}

func Delete(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** DELETE /kubectl")
	queryParams := req.URL.Query()
	userName := queryParams.Get("userName")
	if err := caller.DeleteKubectlResourceByUserName(userName); err != nil {
		util.SetResponse(res, "", err, http.StatusInternalServerError)
	} else {
		util.SetResponse(res, "Delete ["+userName+"] Kubectl Related Resource Success!", nil, http.StatusOK)
	}
}
