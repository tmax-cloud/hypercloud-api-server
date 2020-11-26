package namespaceClaim

import (
	"hypercloud-api-server/util"
	k8sApiCaller "hypercloud-api-server/util/Caller"
	"net/http"
	"strconv"

	claim "github.com/tmax-cloud/hypercloud-go-operator/api/v1alpha1"

	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET/namespaceClaim")
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	labelSelector := queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR)
	klog.Infoln("userId : ", userId)
	klog.Infoln("limit : ", limit)
	klog.Infoln("labelSelector : ", labelSelector)

	var nscList claim.NamespaceClaimList
	nscList = k8sApiCaller.GetAccessibleNSC(userId, labelSelector)

	var status int
	//make OutDO
	if nscList.ResourceVersion != "" {
		status = http.StatusOK
		if len(nscList.Items) > 0 {
			if limit != "" {
				limitInt, _ := strconv.Atoi(limit)
				nscList.Items = nscList.Items[:limitInt]
			}
		}
	} else {
		klog.Infoln(" User [ " + userId + " ] has No Permission to Any NamespaceClaim")
		status = http.StatusForbidden
	}
	util.SetResponse(res, "", nscList, status)
}

func Put(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** PUT/namespaceClaim")

}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** OPTIONS/namespaceClaim")
	out := "**** OPTIONS/namespaceClaim"
	util.SetResponse(res, out, nil, http.StatusOK)
}
