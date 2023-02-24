package clusterUpdateClaim

import (
	"net/http"
	"strconv"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
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
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	reason := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT_REASON)
	admit := queryParams.Get(QUERY_PARAMETER_CLUSTER_CLAIM_ADMIT)

	vars := gmux.Vars(req)
	cucName := vars["clusterupdateclaim"]
	cucNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, admit, cucName, cucNamespace); err != nil {
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

	var cuc *claimsv1alpha1.ClusterUpdateClaim
	if cuc, err = caller.GetClusterUpdateClaim(userId, userGroups, cucName, cucNamespace); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if cuc.Status.Phase != "Awaiting" && cuc.Status.Phase != "Rejected" {
		msg := "ClusterUpdateClaim is not awaiting or rejected phase"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// awaiting과 Rejected 통과
	if err := caller.CheckClusterValid(userId, cuc.Spec.ClusterName, cuc.Namespace); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	if _, err = caller.AdmitClusterUpdateClaim(userId, userGroups, cuc, admitBool, reason); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !admitBool {
		msg := "ClusterUpdateClaim is rejected by admin"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusOK)
		return
	}

	util.SetResponse(res, "ClusterUpdateClaim is approved by admin", cuc, http.StatusOK)
}

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	cucNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	if cucNamespace == "" {
		if clusterUpdateClaimList, err := caller.ListAllClusterUpdateClaims(userId, userGroups); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clusterUpdateClaimList, http.StatusOK)
			return
		}
	} else {
		if clusterUpdateClaimList, err := caller.ListClusterUpdateClaimsByNamespace(userId, userGroups, cucNamespace); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clusterUpdateClaimList, http.StatusOK)
			return
		}
	}

}
