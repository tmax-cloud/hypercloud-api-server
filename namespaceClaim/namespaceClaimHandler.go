package namespaceClaim

import (
	"net/http"
	"strconv"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
)

func Get(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** GET/namespaceClaim")
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	labelSelector := queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	var status int
	klog.V(3).Infoln("userId : ", userId)
	if userId == "" {
		out := "userId is missing"
		status = http.StatusBadRequest
		util.SetResponse(res, out, nil, status)
		return
	}

	klog.V(3).Infoln("limit : ", limit)
	klog.V(3).Infoln("labelSelector : ", labelSelector)

	nscList, err := k8sApiCaller.GetAccessibleNSC(userId, userGroups, labelSelector)
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

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
		klog.V(3).Infoln(" User [ " + userId + " ] has No Permission to Any NamespaceClaim")
		status = http.StatusForbidden
	}
	util.SetResponse(res, "", nscList, status)
}

func Put(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** PUT/namespaceClaim")
	klog.V(3).Infoln(" Namespace Name Duplication Verify Service Start ")
	queryParams := req.URL.Query()
	nsName := queryParams.Get(util.QUERY_PARAMETER_NAMESPACE)
	klog.V(3).Infoln(" Namespace Name : " + nsName)
	var status int
	var out string
	if nsName == "" {
		status = http.StatusBadRequest
		out = "Namespace is missing"
		util.SetResponse(res, out, nil, status)
		return
	}

	namespace, err := k8sApiCaller.GetNamespace(nsName)
	if err != nil && !errors.IsNotFound(err) {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if namespace == nil {
		status = http.StatusOK
		out = "Namespace Duplication verify Success"
	} else {
		status = http.StatusBadRequest
		out = "NameSpace Name is Duplicated"
	}
	util.SetResponse(res, out, nil, status)
}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** OPTIONS/namespaceClaim")
	out := "**** OPTIONS/namespaceClaim"
	util.SetResponse(res, out, nil, http.StatusOK)
}
