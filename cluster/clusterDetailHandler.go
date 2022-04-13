package cluster

import (
	"net/http"
	// "encoding/json"
)

func ListClusterNamespace(res http.ResponseWriter, req *http.Request) {
	// queryParams := req.URL.Query()
	// userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	// userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// accessOnly := queryParams.Get(QUERY_PARAMETER_ACCESS_ONLY)

	// if err := util.StringParameterException(userGroups, userId, accessOnly); err != nil {
	// 	klog.Errorln(err)
	// 	util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
	// 	return
	// }

	// // role 삭제
	// if msg, status = caller.ListNamespaceInRemote(clm, memberId, attribute); status != http.StatusOK {
	// 	util.SetResponse(res, msg, nil, status)
	// 	return
	// }

	// msg = "User [" + memberId + "] is removed from cluster [" + clm.Name + "]"
	// klog.Infoln(msg)
	// util.SetResponse(res, msg, nil, status)
}
