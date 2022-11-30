package event

import (
	"net/http"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"github.com/tmax-cloud/hypercloud-api-server/util/caller"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET /kubectl")
	// TODO
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
