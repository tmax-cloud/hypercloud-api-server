package cluster

import (
	"net/http"
	"strconv"

	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	"k8s.io/klog"
	// "encoding/json"
)

const (
	QUERY_PARAMETER_USER_ID     = "userId"
	QUERY_PARAMETER_LIMIT       = "limit"
	QUERY_PARAMETER_OFFSET      = "offset"
	QUERY_PARAMETER_CLUSTER     = "cluster"
	QUERY_PARAMETER_ACCESS_ONLY = "accessOnly"
	QUERY_PARAMETER_REMOTE_ROLE = "remoteRole"
)

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	accessOnly := queryParams.Get(QUERY_PARAMETER_ACCESS_ONLY)

	if err := util.StringParameterException(userGroups, userId, accessOnly); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	accessOnlyBool, err := strconv.ParseBool(accessOnly)
	if err != nil {
		msg := "AccessOnly parameter has invalid syntax."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := caller.ListCluster(userId, userGroups, accessOnlyBool)
	util.SetResponse(res, msg, clmList, status)
	return
}
