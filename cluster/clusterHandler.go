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
	QUERY_PARAMETER_TARGET_USER  = "targetUser"
	QUERY_PARAMETER_TARGET_GROUP = "targetGroup"
	QUERY_PARAMETER_ACCESS_ONLY  = "accessOnly"
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

	access, err := strconv.ParseBool(queryParams.Get(QUERY_PARAMETER_ACCESS_ONLY))
	if err != nil {
		msg := "Access parameter has invalid syntax."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clmList, msg, status := caller.ListCluster(userId, userGroups, access)
	util.SetResponse(res, msg, clmList, status)
	return
}

func UpdateMemberRole(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	targetUser := queryParams.Get(QUERY_PARAMETER_TARGET_USER)
	targetGroup := queryParams.Get(QUERY_PARAMETER_TARGET_GROUP)
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

	if remoteRole == "" {
		msg := "RemoteRole is empty."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if clm.Status.Ready == false {
		msg := "Cannot modify cluster member in not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if targetUser != "" && targetGroup == "" {
		if _, ok := clm.Status.Members[targetUser]; ok {
			clm.Status.Members[targetUser] = remoteRole
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.RemoveRoleFromRemote(updatedClm, targetUser, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateRoleInRemote(updatedClm, targetUser, remoteRole, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "User [" + targetUser + "] role is updated to [" + remoteRole + "] in cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "User [" + targetUser + "] is not a member of cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if targetGroup != "" && targetUser == "" {
		if _, ok := clm.Status.Groups[targetGroup]; ok {
			clm.Status.Groups[targetGroup] = remoteRole
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.RemoveRoleFromRemote(updatedClm, targetGroup, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateRoleInRemote(updatedClm, targetGroup, remoteRole, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "Group [" + targetGroup + "] role is updated to [" + remoteRole + "] in cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "Group [" + targetGroup + "] is not a member of cluster [" + clm.Name + "]"
			klog.Errorln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if targetGroup == "" && targetUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot update role for user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}

// func InviteMember(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
// 	//
// 	targetUser := queryParams.Get(QUERY_PARAMETER_TARGET_USER)
// 	targetGroup := queryParams.Get(QUERY_PARAMETER_TARGET_GROUP)
// 	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

// 	if userId == "" {
// 		msg := "UserId is empty."
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	if cluster == "" {
// 		msg := "cluster is empty."
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	if remoteRole == "" {
// 		msg := "RemoteRole is empty."
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
// 	if clm == nil {
// 		util.SetResponse(res, msg, nil, status)
// 		return
// 	}
// 	if clm.Status.Ready == false {
// 		msg := "Cannot invite member to cluster in not ready status"
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	if targetUser != "" && targetGroup == "" {
// 		if _, ok := clm.Status.Members[targetUser]; !ok {
// 			// if !util.Contains(clm.Status.Members, targetUser) {
// 			if clm.Status.Members == nil {
// 				clm.Status.Members = map[string]string{}
// 			}
// 			clm.Status.Members[targetUser] = remoteRole
// 			// clm.Status.Members = append(clm.Status.Members, targetUser)
// 			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
// 			if updatedClm != nil {
// 				if msg, status := caller.CreateCLMRole(updatedClm, targetUser, false); status != http.StatusOK {
// 					util.SetResponse(res, msg, nil, status)
// 					return
// 				}
// 				if msg, status := caller.CreateRoleInRemote(updatedClm, targetUser, remoteRole, false); status != http.StatusOK {
// 					util.SetResponse(res, msg, nil, status)
// 					return
// 				}
// 				msg = "User [" + targetUser + "] is added to cluster [" + clm.Name + "]"
// 				klog.Infoln(msg)
// 				util.SetResponse(res, msg, updatedClm, status)
// 				return
// 			} else {
// 				klog.Errorln(msg)
// 				util.SetResponse(res, msg, nil, status)
// 				return
// 			}
// 		} else {
// 			msg := "User [" + targetUser + "] is already added to cluster [" + clm.Name + "]"
// 			klog.Infoln(msg)
// 			util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 			return
// 		}
// 	} else if targetGroup != "" && targetUser == "" {
// 		if _, ok := clm.Status.Groups[targetGroup]; !ok {
// 			if clm.Status.Groups == nil {
// 				clm.Status.Groups = map[string]string{}
// 			}
// 			clm.Status.Groups[targetGroup] = remoteRole
// 			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
// 			if updatedClm != nil {
// 				if msg, status := caller.CreateCLMRole(updatedClm, targetGroup, true); status != http.StatusOK {
// 					util.SetResponse(res, msg, nil, status)
// 					return
// 				}
// 				if msg, status := caller.CreateRoleInRemote(updatedClm, targetGroup, remoteRole, true); status != http.StatusOK {
// 					util.SetResponse(res, msg, nil, status)
// 					return
// 				}
// 				msg = "Group [" + targetGroup + "] is added to cluster [" + clm.Name + "]"
// 				klog.Infoln(msg)
// 				util.SetResponse(res, msg, updatedClm, status)
// 				return
// 			} else {
// 				klog.Errorln(msg)
// 				util.SetResponse(res, msg, nil, status)
// 				return
// 			}
// 		} else {
// 			msg := "Group [" + targetGroup + "] is already added to cluster [" + clm.Name + "]"
// 			klog.Infoln(msg)
// 			util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 			return
// 		}
// 	} else if targetGroup == "" && targetUser == "" {
// 		msg := "Both user and group is empty in request."
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	} else {
// 		msg := "Cannot add user and group at the same time."
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}
// }

func RemoveMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	targetUser := queryParams.Get(QUERY_PARAMETER_TARGET_USER)
	targetGroup := queryParams.Get(QUERY_PARAMETER_TARGET_GROUP)

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
	if clm.Status.Ready == false {
		msg := "Cannot remove member from cluster in not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if targetUser != "" && targetGroup == "" {
		if _, ok := clm.Status.Members[targetUser]; ok {
			// if util.Contains(clm.Status.Members, targetUser) {
			// delete(clm.Status.Members, targetUser)
			// clm.Status.Members = util.Remove(clm.Status.Members, targetUser)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status = caller.DeleteCLMRole(updatedClm, targetUser, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status = caller.RemoveRoleFromRemote(updatedClm, targetUser, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "User [" + targetUser + "] is removed from cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "User [" + targetUser + "] is already removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if targetGroup != "" && targetUser == "" {
		if _, ok := clm.Status.Groups[targetGroup]; ok {
			// if util.Contains(clm.Status.Groups, targetGroup) {
			// clm.Status.Groups = util.Remove(clm.Status.Groups, targetGroup)
			// delete(clm.Status.Groups, targetGroup)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status = caller.DeleteCLMRole(updatedClm, targetGroup, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status = caller.RemoveRoleFromRemote(updatedClm, targetGroup, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "Group [" + targetGroup + "] is removed from cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "Group [" + targetGroup + "] is already removed from cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if targetGroup == "" && targetUser == "" {
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
