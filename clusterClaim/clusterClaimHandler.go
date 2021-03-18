package clusterClaim

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	claimsv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/claim/v1alpha1"
	"k8s.io/klog"
)

const (
	QUERY_PARAMETER_USER_ID                    = "userId"
	QUERY_PARAMETER_USER_NAME                  = "userName"
	QUERY_PARAMETER_LABEL_SELECTOR             = "labelSelector"
	QUERY_PARAMETER_LIMIT                      = "limit"
	QUERY_PARAMETER_OFFSET                     = "offset"
	QUERY_PARAMETER_CLUSTER_CLAIM              = "clusterClaim"
	QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT        = "admit"
	QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON = "reason"
)

func Post(res http.ResponseWriter, req *http.Request) {
	var body []byte
	if req.Body != nil {
		if data, err := ioutil.ReadAll(req.Body); err == nil {
			body = data
		}
	}

	cc := &claimsv1alpha1.ClusterClaim{}
	if err := json.Unmarshal(body, cc); err != nil {
		klog.Error(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	result, msg, status := caller.CreateClusterClaim(userId, userGroups, cc)
	if cc == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	msg, status = caller.CreateCCRole(userId, userGroups, result)
	util.SetResponse(res, msg, result, status)
	return
}

func Put(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userName := queryParams.Get(QUERY_PARAMETER_USER_NAME)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	reason := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON)
	admit := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT)

	vars := gmux.Vars(req)
	clusterClaim := vars["clusterclaim"]

	if err := util.StringParameterException(userGroups, userId, admit, reason); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	admitBool, err := strconv.ParseBool(admit)
	if err != nil {
		msg := "Admit parameter has invalid syntax."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	cc, msg, status := caller.GetClusterClaim(userId, userGroups, clusterClaim)
	if cc == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if cc.Status.Phase != "Awaiting" {
		msg = "ClusterClaim is already admitted or rejected by admin"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cc.Spec.ClusterName
	clusterMember.Role = "admin"
	clusterMember.MemberId = userId
	clusterMember.MemberName = userName
	clusterMember.Attribute = "user"
	clusterMember.Status = "owner"
	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	updatedClusterClaim, msg, status := caller.AdmitClusterClaim(userId, userGroups, cc, admitBool, reason)
	if updatedClusterClaim == nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		if err := clusterDataFactory.Delete(clusterMember); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		}
		return
	}

	util.SetResponse(res, msg, updatedClusterClaim, status)
	return
}

func List(res http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	// limit, err := strconv.Atoi(queryParams.Get(QUERY_PARAMETER_LIMIT))

	// if err != nil {
	// 	out := "Limit parameter has invalid syntax."
	// 	util.SetResponse(res, out, nil, http.StatusBadRequest)
	// 	return
	// }

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterClaimList, msg, status := caller.ListAccessibleClusterClaims(userId, userGroups)

	util.SetResponse(res, msg, clusterClaimList, status)
	return
}
