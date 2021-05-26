package cluster

import (
	"net/http"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	"k8s.io/klog"
	// "encoding/json"
)

const (
	// QUERY_PARAMETER_USER_NAME   = "Name"
	QUERY_PARAMETER_USER_ID     = "userId"
	QUERY_PARAMETER_USER_NAME   = "userName"
	QUERY_PARAMETER_LIMIT       = "limit"
	QUERY_PARAMETER_OFFSET      = "offset"
	QUERY_PARAMETER_CLUSTER     = "cluster"
	QUERY_PARAMETER_ACCESS_ONLY = "accessOnly"
	QUERY_PARAMETER_REMOTE_ROLE = "remoteRole"
	QUERY_PARAMETER_MEMBER_NAME = "memberName"
)

func ListPage(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// accessOnly := queryParams.Get(QUERY_PARAMETER_ACCESS_ONLY)
	vars := gmux.Vars(req)
	clusterNamespace := vars["namespaces"]

	if err := util.StringParameterException(userGroups, userId, clusterNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterClaimList, msg, status := caller.ListClusterInNamespace(userId, userGroups, clusterNamespace)
	util.SetResponse(res, msg, clusterClaimList, status)
	return

}

func ListLNB(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := caller.ListAccesibleCluster(userId, userGroups)
	util.SetResponse(res, msg, clmList, status)
	return

}
