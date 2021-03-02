package cluster

// "encoding/json"
import (
	"net/http"

	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	"k8s.io/klog"
	// "encoding/json"
)

type invitationInfo struct {
	id        int64
	cluster   string
	member    string
	attribute string
	status    string
}

func InviteMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
	targetUser := queryParams.Get(QUERY_PARAMETER_TARGET_USER)
	targetGroup := queryParams.Get(QUERY_PARAMETER_TARGET_GROUP)
	// remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

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

	// if remoteRole == "" {
	// 	msg := "RemoteRole is empty."
	// 	klog.Infoln(msg)
	// 	util.SetResponse(res, msg, nil, http.StatusBadRequest)
	// 	return
	// }

	invitation := invitationInfo{}
	invitation.cluster = cluster

	if targetUser != "" && targetGroup == "" {
		invitation.member = targetUser
		invitation.attribute = "user"
	} else if targetGroup != "" && targetUser == "" {
		invitation.member = targetGroup
		invitation.attribute = "group"
	} else if targetGroup == "" && targetUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot invite user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// 권한 체크
	// caller
	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "update")
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !sarResult.Status.Allowed {
		msg := " User [ " + userId + " ] has No ClusterManagers Update Role"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if err := insert(invitation); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	//성공했다고 msg주자
	return
	// 메일 보내고
	// if err := util.SendEmail(from string, to []string, subject string, body string); err != nil{
	// 	klog.Errorln(err)
	// 	util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
	// 	return
	// }
}

func AcceptInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
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
		msg := "Cannot invite member to cluster in not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// db에 있는지 확인...
	//

	if targetUser != "" && targetGroup == "" {
		//
		if invitation, err := get(cluster, targetUser, "user"); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else if invitation == nil {
			msg = "Couldn't find that cluster invitation"
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
		if _, ok := clm.Status.Members[targetUser]; !ok {
			// if !util.Contains(clm.Status.Members, targetUser) {
			if clm.Status.Members == nil {
				clm.Status.Members = map[string]string{}
			}
			clm.Status.Members[targetUser] = remoteRole
			// clm.Status.Members = append(clm.Status.Members, targetUser)
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.CreateCLMRole(updatedClm, targetUser, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateRoleInRemote(updatedClm, targetUser, remoteRole, false); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "User [" + targetUser + "] is added to cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "User [" + targetUser + "] is already added to cluster [" + clm.Name + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
	} else if targetGroup != "" && targetUser == "" {
		if invitation, err := get(cluster, targetGroup, "group"); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else if invitation == nil {
			msg = "Couldn't find that cluster invitation"
			util.SetResponse(res, msg, nil, http.StatusBadRequest)
			return
		}
		if _, ok := clm.Status.Groups[targetGroup]; !ok {
			if clm.Status.Groups == nil {
				clm.Status.Groups = map[string]string{}
			}
			clm.Status.Groups[targetGroup] = remoteRole
			updatedClm, msg, status := caller.UpdateClusterManager(userId, userGroups, clm)
			if updatedClm != nil {
				if msg, status := caller.CreateCLMRole(updatedClm, targetGroup, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				if msg, status := caller.CreateRoleInRemote(updatedClm, targetGroup, remoteRole, true); status != http.StatusOK {
					util.SetResponse(res, msg, nil, status)
					return
				}
				msg = "Group [" + targetGroup + "] is added to cluster [" + clm.Name + "]"
				klog.Infoln(msg)
				util.SetResponse(res, msg, updatedClm, status)
				return
			} else {
				klog.Errorln(msg)
				util.SetResponse(res, msg, nil, status)
				return
			}
		} else {
			msg := "Group [" + targetGroup + "] is already added to cluster [" + clm.Name + "]"
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

func DeclineInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	// userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
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

	invitation := invitationInfo{}
	invitation.cluster = cluster

	if targetUser != "" && targetGroup == "" {
		invitation.member = targetUser
		invitation.attribute = "user"
		if err := delete(invitation); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		}
	} else if targetGroup != "" && targetUser == "" {
		invitation.member = targetGroup
		invitation.attribute = "group"
		if err := delete(invitation); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		}
	} else if targetGroup == "" && targetUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot delete user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}

func GetInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	// userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)
	//
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

	invitation := invitationInfo{}
	invitation.cluster = cluster

	if targetUser != "" && targetGroup == "" {
		invitation.member = targetUser
		invitation.attribute = "user"
		if err := delete(invitation); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		}
	} else if targetGroup != "" && targetUser == "" {
		invitation.member = targetGroup
		invitation.attribute = "group"
		if err := delete(invitation); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		}
	} else if targetGroup == "" && targetUser == "" {
		msg := "Both user and group is empty in request."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else {
		msg := "Cannot delete user and group at the same time."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}
}
