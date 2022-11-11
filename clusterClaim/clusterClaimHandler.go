package clusterClaim

import (
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
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	admitBool, err := strconv.ParseBool(admit)
	if err != nil {
		msg := "Admit parameter has invalid syntax."
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	var cc *claimsv1alpha1.ClusterClaim
	if cc, err = caller.GetClusterClaim(userId, userGroups, clusterClaimName, clusterClaimNamespace); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !(cc.Status.Phase == "Awaiting" || cc.Status.Phase == "Rejected"){
		msg := "ClusterClaim is already admitted by admin"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// name duplicate
	exist, err := caller.CheckClusterManagerDuplication(cc.Spec.ClusterName, clusterClaimNamespace)
	if err != nil {
		klog.V(1).Infoln(err.Error())
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if exist {
		msg := "Cluster [" + cc.Spec.ClusterName + "] is already existed."
		klog.V(3).Infoln(msg)
		util.SetResponse(res, "Cluster ["+cc.Spec.ClusterName+"] is already existed.", nil, http.StatusBadRequest)
		return
	}

	var updatedClusterClaim *claimsv1alpha1.ClusterClaim
	if updatedClusterClaim, err = caller.AdmitClusterClaim(userId, userGroups, cc, admitBool, reason); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	if updatedClusterClaim.Status.Phase == "Rejected" {
		msg := "ClusterClaim is rejected by admin"
		klog.V(3).Infoln(msg)
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

	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	util.SetResponse(res, "Success", updatedClusterClaim, http.StatusOK)
}

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	clusterClaimNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}
	// var clusterClaimList *claimsv1alpha1.ClusterClaimList
	if clusterClaimNamespace == "" {
		if clusterClaimList, err := caller.ListAllClusterClaims(userId, userGroups); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clusterClaimList, http.StatusOK)
			return
		}
	} else {
		if clusterClaimList, err := caller.ListAccessibleClusterClaims(userId, userGroups, clusterClaimNamespace); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clusterClaimList, http.StatusOK)
			return
		}
	}

}
