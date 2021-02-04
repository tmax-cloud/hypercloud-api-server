package namespaceClaim

import (
	"net/http"
	"strconv"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	claim "github.com/tmax-cloud/hypercloud-go-operator/api/v1alpha1"

	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** GET/namespaceClaim")
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	labelSelector := queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	var status int
	klog.Infoln("userId : ", userId)
	if userId == "" {
		out := "userId is missing"
		status = http.StatusBadRequest
		util.SetResponse(res, out, nil, status)
		return
	}

	klog.Infoln("limit : ", limit)
	klog.Infoln("labelSelector : ", labelSelector)

	var nscList claim.NamespaceClaimList
	nscList = k8sApiCaller.GetAccessibleNSC(userId, userGroups, labelSelector)

	//make OutDO
	if nscList.ResourceVersion != "" {
		status = http.StatusOK
		if len(nscList.Items) > 0 {
			if limit != "" {
				limitInt, _ := strconv.Atoi(limit)
				if len(nscList.Items) < limitInt {
					limitInt = len(nscList.Items)
				}
				nscList.Items = nscList.Items[:limitInt]
			}
		}
	} else {
		klog.Infoln(" User [ " + userId + " ] has No Permission to Any NamespaceClaim")
		status = http.StatusForbidden
	}
	util.SetResponse(res, "", nscList, status)
	return
}

func Put(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** PUT/namespaceClaim")
	klog.Infoln(" Namespace Name Duplication Verify Service Start ")
	queryParams := req.URL.Query()
	nsName := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	klog.Infoln(" Namespace Name : " + nsName)
	var status int
	var out string
	if nsName == "" {
		status = http.StatusBadRequest
		out = "Namespace is missing"
		util.SetResponse(res, out, nil, status)
		return
	}
	namespace := k8sApiCaller.GetNamespace(nsName)
	if namespace == nil {
		status = http.StatusOK
		out = "Namespace Duplication verify Success"
	} else {
		status = http.StatusBadRequest
		out = "NameSpace Name is Duplicated"
	}
	util.SetResponse(res, out, nil, status)
	return
}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** OPTIONS/namespaceClaim")
	out := "**** OPTIONS/namespaceClaim"
	util.SetResponse(res, out, nil, http.StatusOK)
	return
}
