package clusterClaim

import (
	"net/http"
	"strconv"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	"k8s.io/klog"
)

const (
	QUERY_PARAMETER_USER_ID                    = "userId"
	QUERY_PARAMETER_MEMBER_NAME                = "memberName"
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
	memberName := queryParams.Get(QUERY_PARAMETER_MEMBER_NAME)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	reason := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON)
	admit := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT)

	vars := gmux.Vars(req)
	clusterClaimName := vars["clusterclaim"]
	clusterClaimNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, admit, memberName, clusterClaimName, clusterClaimNamespace); err != nil {
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

	cc, msg, status := caller.GetClusterClaim(userId, userGroups, clusterClaimName, clusterClaimNamespace)
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

	// name duplicate
	exist, err := caller.CheckClusterManagerDupliation(userId, userGroups, clusterClaimName, clusterClaimNamespace)
	if err != nil {
		util.SetResponse(res, "", nil, http.StatusInternalServerError)
		return
	}

	if exist {
		util.SetResponse(res, "Cluster ["+cc.Spec.ClusterName+"] is existed.", nil, http.StatusBadRequest)
		return
	}

	updatedClusterClaim, msg, status := caller.AdmitClusterClaim(userId, userGroups, cc, admitBool, reason)
	if updatedClusterClaim == nil {
		klog.Errorln(msg)
		util.SetResponse(res, msg, nil, status)
		return
	}

	if updatedClusterClaim.Status.Phase == "Rejected" {
		msg = "ClusterClaim is rejected by admin"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusOK)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = cc.Namespace
	clusterMember.Cluster = cc.Spec.ClusterName
	clusterMember.Role = "admin"
	clusterMember.MemberId = cc.Annotations["creator"]
	clusterMember.MemberName = memberName
	clusterMember.Attribute = "user"
	clusterMember.Status = "owner"

	clm, msg, status := caller.CreateClusterManager(updatedClusterClaim)
	if clm == nil {
		klog.Errorln(msg)
		util.SetResponse(res, msg, nil, http.StatusInternalServerError)
		return
	}

	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	util.SetResponse(res, "Succes", updatedClusterClaim, http.StatusOK)
	return
}

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	clusterClaimNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}
	if clusterClaimNamespace == "" {
		clusterClaimList, msg, status := caller.ListAllClusterClaims(userId, userGroups)
		util.SetResponse(res, msg, clusterClaimList, status)
		return
	} else {
		clusterClaimList, msg, status := caller.ListAccessibleClusterClaims(userId, userGroups, clusterClaimNamespace)
		util.SetResponse(res, msg, clusterClaimList, status)
		return
	}
}
