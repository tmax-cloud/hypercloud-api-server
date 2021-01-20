package cluster

import (
	"net/http"

	util "github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/Caller"

	"k8s.io/klog"
	// "encoding/json"
)

const (
	QUERY_PARAMETER_USER_ID     = "userId"
	QUERY_PARAMETER_LIMIT       = "limit"
	QUERY_PARAMETER_OFFSET      = "offset"
	QUERY_PARAMETER_CLUSTER     = "cluster"
	QUERY_PARAMETER_NEW_USER    = "newUser"
	QUERY_PARAMETER_DELETE_USER = "deleteUser"
)

func Put(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	clusterName := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	newUsers := queryParams[QUERY_PARAMETER_NEW_USER]
	deletedUsers := queryParams[QUERY_PARAMETER_DELETE_USER]

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if clusterName == "" {
		msg := "ClusterName is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clm, msg, status := k8sApiCaller.GetCluster(userId, clusterName)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if len(newUsers) != 0 && len(deletedUsers) == 0 {
		var newMembers []string
		for _, newUser := range newUsers {
			if !util.Contains(clm.Status.Members, newUser) {
				newMembers = append(newMembers, newUser)
			}
		}
		updatedClm, msg, status := k8sApiCaller.AddMembers(userId, clm, newMembers)
		msg, status = k8sApiCaller.CreateCLMRole(updatedClm, newMembers)
		util.SetResponse(res, msg, updatedClm, status)
		return
	} else if len(deletedUsers) != 0 && len(newUsers) == 0 {
		// var deletedMembers []string
		updatedClm, msg, status := k8sApiCaller.DeleteMembers(userId, clm, deletedUsers)
		msg, status = k8sApiCaller.DeleteCLMRole(updatedClm, deletedUsers)
		util.SetResponse(res, msg, updatedClm, status)
		return
	} else if len(deletedUsers) == 0 && len(newUsers) == 0 {
		msg := "Both added and deleted user is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot add & delete members at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := k8sApiCaller.ListCluster(userId)

	util.SetResponse(res, msg, clmList, status)
	return
}

func ListOwner(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := k8sApiCaller.ListOwnerCluster(userId)
	util.SetResponse(res, msg, clmList, status)
	return
}

func ListMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := k8sApiCaller.ListMemberCluster(userId)
	util.SetResponse(res, msg, clmList, status)
	return
}
