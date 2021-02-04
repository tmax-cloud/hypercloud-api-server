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
	QUERY_PARAMETER_USER_ID      = "userId"
	QUERY_PARAMETER_LIMIT        = "limit"
	QUERY_PARAMETER_OFFSET       = "offset"
	QUERY_PARAMETER_CLUSTER      = "cluster"
	QUERY_PARAMETER_INVITE_USER  = "inviteUser"
	QUERY_PARAMETER_INVITE_GROUP = "inviteGroup"
	QUERY_PARAMETER_REMOVE_USER  = "removeUser"
	QUERY_PARAMETER_REMOVE_GROUP = "removeGroup"
	QUERY_PARAMETER_ACCESS       = "accessible"
	QUERY_PARAMETER_REMOTE_ROLE  = "remoteRole"
)

func List(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	accessible, err := strconv.ParseBool(queryParams.Get(QUERY_PARAMETER_ACCESS))
	if err != nil {
		msg := "Access parameter has invalid syntax."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := caller.ListCluster(userId, userGroups, accessible)
	util.SetResponse(res, msg, clmList, status)

	return
}

func InviteMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
	inviteUser := queryParams.Get(QUERY_PARAMETER_INVITE_USER)
	inviteGroup := queryParams.Get(QUERY_PARAMETER_INVITE_GROUP)
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if cluster == "" {
		msg := "cluster is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if inviteUser != "" && inviteGroup == "" {
		if !util.Contains(clm.Status.Members, inviteUser) {
			clm.Status.Members = append(clm.Status.Members, inviteUser)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.CreateCLMRole(updatedClm, inviteUser, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateSubjectRolebinding(updatedClm, inviteUser, remoteRole, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
			}
			msg = "User [" + inviteUser + "] is added to cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, updatedClm, status)
			return
		} else {
			msg := "User [" + inviteUser + "] is already added to cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if inviteGroup != "" && inviteUser == "" {
		if !util.Contains(clm.Status.Groups, inviteGroup) {
			clm.Status.Groups = append(clm.Status.Groups, inviteGroup)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.CreateCLMRole(updatedClm, inviteGroup, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateSubjectRolebinding(updatedClm, inviteGroup, remoteRole, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
			}
			msg = "Group [" + inviteUser + "] is added to cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, updatedClm, status)
			return
		} else {
			msg := "Group [" + inviteGroup + "] is already added to cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if inviteGroup == "" && inviteUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot add user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}

func RemoveMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
	removeUser := queryParams.Get(QUERY_PARAMETER_REMOVE_USER)
	removeGroup := queryParams.Get(QUERY_PARAMETER_REMOVE_GROUP)

	if userId == "" {
		msg := "UserId is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if cluster == "" {
		msg := "cluster is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if removeUser != "" && removeGroup == "" {
		if util.Contains(clm.Status.Members, removeUser) {
			clm.Status.Members = util.Remove(clm.Status.Members, removeUser)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status = caller.DeleteCLMRole(updatedClm, removeUser); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status = caller.RemoveSubjectRolebinding(updatedClm, removeUser); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
			}
			msg = "User [" + removeUser + "] is removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, updatedClm, status)
			return
		} else {
			msg := "User [" + removeUser + "] is already removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if removeGroup != "" && removeUser == "" {
		if util.Contains(clm.Status.Groups, removeGroup) {
			clm.Status.Groups = util.Remove(clm.Status.Groups, removeGroup)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if updatedClm != nil {
					if msg, status = caller.DeleteCLMRole(updatedClm, removeUser); status != http.StatusOK {
						util.SetResponse(res, msg, nil, status)
						return
					}
					if msg, status = caller.RemoveSubjectRolebinding(updatedClm, removeUser); status != http.StatusOK {
						util.SetResponse(res, msg, nil, status)
						return
					}
				}
			}
			msg = "Group [" + removeGroup + "] is removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, updatedClm, status)
			return
		} else {
			msg := "Group [" + removeGroup + "] is already removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if removeGroup == "" && removeUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot add user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}
