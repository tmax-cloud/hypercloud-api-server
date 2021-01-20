package claim

import (
	// _ "github.com/go-sql-driver/mysql"

	"net/http"
	"strconv"

	// "encoding/json"
	util "github.com/tmax-cloud/hypercloud-multi-api-server/util"
	"k8s.io/klog"
)

const (
	QUERY_PARAMETER_USER_ID                    = "userId"
	QUERY_PARAMETER_LABEL_SELECTOR             = "labelSelector"
	QUERY_PARAMETER_LIMIT                      = "limit"
	QUERY_PARAMETER_OFFSET                     = "offset"
	QUERY_PARAMETER_CLUSTER_CLAIM              = "clusterClaim"
	QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT        = "admit"
	QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON = "reason"
)

func Put(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	clusterClaimName := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM)
	admit, err := strconv.ParseBool(queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT))
	reason := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON)

	if err != nil {
		msg := "Admit parameter has invalid syntax."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterClaim, msg, status := util.GetClusterClaim(userId, clusterClaimName)
	if clusterClaim == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if clusterClaim.Status.Phase != "Awaiting" {
		msg = "ClusterClaim is already admitted or rejected by admin"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	updatedClusterClaim, msg, status := util.AdmitClusterClaim(userId, clusterClaim, admit, reason)
	if updatedClusterClaim == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	util.SetResponse(res, msg, updatedClusterClaim, status)
	return
}

func List(res http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	// limit, err := strconv.Atoi(queryParams.Get(QUERY_PARAMETER_LIMIT))

	// if err != nil {
	// 	out := "Limit parameter has invalid syntax."
	// 	util.SetResponse(res, out, nil, http.StatusBadRequest)
	// 	return
	// }

	if userId == "" {
		msg := "UserId is empty."
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
	// // var statusCode int
	clusterClaimList, msg, status := util.ListAccessibleClusterClaims(userId)

	// if clusterClaimList.ResourceVersion != "" {
	// 	status = http.StatusOK
	// 	if len(clusterClaimList.Items) > 0 {
	// 		if limit != "" {
	// 			limitInt, _ := strconv.Atoi(limit)
	// 			clusterClaimList.Items = clusterClaimList.Items[:limitInt]
	// 		}
	// 	}
	// } else {
	// 	klog.Infoln(" User [ " + userId + " ] has No Permission to Any NamespaceClaim")
	// 	status = http.StatusForbidden
	// }
	util.SetResponse(res, msg, clusterClaimList, status)
	return
}
