package event

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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
	queryParams := req.URL.Query()
	userName := queryParams.Get("userName")
	klog.V(3).Infoln("userName =", userName)
	sleepTime := os.Getenv("KUBECTL_TIMEOUT")
	if len(sleepTime) == 0 || sleepTime == "{KUBECTL_TIMEOUT}" {
		sleepTime = "21600" // 6 hours
	}

	pods, exist := caller.GetPodListByLabel(util.HYPERCLOUD_KUBECTL_LABEL_KEY+"="+util.HYPERCLOUD_KUBECTL_LABEL_VALUE, util.HYPERCLOUD_KUBECTL_NAMESPACE)
	if !exist {
		util.SetResponse(res, "No pods in hypercloud-kubectl namespace", nil, http.StatusOK)
		return
	}

	kubectlName := util.HYPERCLOUD_KUBECTL_PREFIX + caller.ParseUserName(userName)
	for _, p := range pods.Items {
		if p.Name == kubectlName {
			createTime := p.CreationTimestamp.Time
			currentTime := time.Now()
			spentTime := currentTime.Sub(createTime).Seconds()
			var remain float64
			if sleepTimeFloat, err := strconv.ParseFloat(sleepTime, 64); err != nil {
				klog.V(1).Infoln(err)
			} else {
				remain = sleepTimeFloat - spentTime
			}
			klog.V(3).Infoln(remain)
			remainTime := fmt.Sprintf("%f", remain)
			klog.V(3).Infoln(remainTime)
			kl := KubectlInfo{
				Image:   p.Spec.Containers[0].Image,
				Timeout: remainTime,
			}
			util.SetResponse(res, "", kl, http.StatusOK)
			return
		}
	}

	// 콘솔에서 POST, GET 콜을 동시에 요청하기 때문에,
	// 최초 요청시 GET 콜 시 pod가 생성되기 이전이라 정보를 가져올 수 없음
	// 따라서 pod가 생성된다고 가정하여 default Timeout 시간 반환
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
	klog.V(3).Infoln("userName =", userName)
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
	klog.V(3).Infoln("userName =", userName)
	if err := caller.DeleteKubectlResourceByUserName(userName); err != nil {
		util.SetResponse(res, "", err, http.StatusInternalServerError)
	} else {
		util.SetResponse(res, "Delete ["+userName+"] Kubectl Related Resource Success!", nil, http.StatusOK)
	}
}
