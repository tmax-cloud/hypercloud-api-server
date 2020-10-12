package namespace

import (
	"hypercloud-api-server/util"
	k8sApiCaller "hypercloud-api-server/util/Caller"
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"net/http"
	"strconv"
)

func Get(res http.ResponseWriter, req *http.Request)  {
	klog.Infoln("**** GET/namespace");
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)
	limit := queryParams.Get(util.QUERY_PARAMETER_LIMIT)
	labelSelector := queryParams.Get(util.QUERY_PARAMETER_LABEL_SELECTOR)
	klog.Infoln("userId : ", userId)
	klog.Infoln("limit : ", limit)
	klog.Infoln("labelSelector : ", labelSelector)

	var nsList v1.NamespaceList
	nsList = k8sApiCaller.GetAccessibleNS(userId, labelSelector)

	var status int
	//make OutDO
	if nsList.ResourceVersion != ""{
		status = http.StatusOK
		if len(nsList.Items) > 0 {
			if limit != ""{
				limitInt, _ := strconv.Atoi(limit)
				nsList.Items = nsList.Items [:limitInt]
			}
		}
	} else {
		status = http.StatusForbidden
	}
	util.SetResponse(res, "", nsList, status)
}

func Put(res http.ResponseWriter, req *http.Request)  {
	klog.Infoln("**** PUT/namespace");

}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.Infoln("**** OPTIONS/namespace");
	out := "**** OPTIONS/namespace"
	util.SetResponse(res, out, nil, http.StatusOK)
}