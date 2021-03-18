package cluster

// "encoding/json"
import (
	"net/http"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	"k8s.io/klog"
	// "encoding/json"
)

func ListClusterMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	cluster := vars["cluster"]

	if err := util.StringParameterException(userGroups, userId, cluster); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	// 있는 cluster인지 확인
	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListClusterMember(cluster)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	msg = "List cluster success"
	klog.Infoln(msg)
	util.SetResponse(res, msg, clusterMemberList, http.StatusOK)
	return
}

func RemoveMember(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	attribute := vars["attribute"]
	memberId := vars["member"]

	if err := util.StringParameterException(userGroups, userId, cluster, attribute, memberId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
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

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	clusterMember.MemberId = memberId
	clusterMember.Attribute = attribute
	clusterMember.Status = "invited"

	// db에서 삭제
	if err := clusterDataFactory.Delete(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// role 삭제
	if msg, status = caller.DeleteCLMRole(clm, memberId, attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if msg, status = caller.RemoveRoleFromRemote(clm, memberId, attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}
	msg = "User [" + memberId + "] is removed from cluster [" + clm.Name + "]"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, status)
	return
}

func UpdateMemberRole(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	attribute := vars["attribute"]
	memberId := vars["member"]

	if err := util.StringParameterException(userGroups, userId, cluster, attribute, memberId, remoteRole); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
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

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	clusterMember.MemberId = memberId
	clusterMember.Role = remoteRole
	clusterMember.Attribute = attribute
	clusterMember.Status = "invited"

	// db에서 role update
	if err := clusterDataFactory.UpdateRole(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// role 삭제 후 재 생성
	if msg, status := caller.RemoveRoleFromRemote(clm, memberId, attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if msg, status := caller.CreateRoleInRemote(clm, memberId, remoteRole, attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}
	msg = attribute + " [" + memberId + "] role is updated to [" + remoteRole + "] in cluster [" + clm.Name + "]"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, status)
	return

}

// func InviteUser(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
// 	vars := gmux.Vars(req)
// 	cluster := vars["cluster"]
// 	member := vars["member"]

// 	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, member); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}

// 	clusterMember := util.ClusterMemberInfo{}
// 	clusterMember.Cluster = cluster
// 	clusterMember.Role = remoteRole
// 	clusterMember.Member = member
// 	clusterMember.Attribute = "user"
// 	clusterMember.Status = "pending"

// 	// db에 이미 있는지 확인
// 	count, err := clusterDataFactory.GetPendingUser(clusterMember)
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	} else if count != 0 {
// 		msg := "Member is already invited in cluster"
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "update")
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}

// 	if !sarResult.Status.Allowed {
// 		msg := " User [ " + userId + " ] is not a owner of cluster. Cannot invtie member"
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	// create Token
// 	token, err := util.CreateToken(clusterMember)

// 	// insert db
// 	if err := clusterDataFactory.Insert(clusterMember); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}

// 	// send mail
// 	// body를 const에서 읽어서 replace해주고 sendMail에 body는 바로 넣자

// 	to := []string{"sangwon9377@naver.com"}
// 	from := "no-reply-tc@tmax.co.kr"
// 	subject := " 신청해주신 Trial NameSpace 만료 안내 "
// 	bodyParameter := map[string]string{}
// 	bodyParameter["%%TO%%"] = to[0]
// 	bodyParameter["%%FROM%%"] = from
// 	bodyParameter["%%CLUSTER%%"] = cluster
// 	bodyParameter["%%IP%%"] = "192.168.6.147"
// 	bodyParameter["%%TOKEN%%"] = token

// 	if err := util.SendEmail("no-reply-tc@tmax.co.kr", to, subject, bodyParameter); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		if err := clusterDataFactory.Delete(clusterMember); err != nil {
// 			klog.Errorln(err)
// 			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		}
// 		return
// 	}

// 	msg := "User inivtation is successed"
// 	klog.Infoln(msg)
// 	util.SetResponse(res, msg, nil, http.StatusOK)
// 	return
// }

// func InviteGroup(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

// 	vars := gmux.Vars(req)
// 	cluster := vars["cluster"]
// 	member := vars["member"]

// 	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, member); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}

// 	clusterMember := util.ClusterMemberInfo{}
// 	clusterMember.Cluster = cluster
// 	clusterMember.Role = remoteRole
// 	clusterMember.Member = member
// 	clusterMember.Attribute = "group"
// 	clusterMember.Status = "invited"

// 	// db에 이미 있는지 확인
// 	count, err := clusterDataFactory.GetInvitedGroup(clusterMember)
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	} else if count != 0 {
// 		msg := "Group is already member of cluster"
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	// cluster ready 인지 확인
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

// 	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "update")
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}

// 	if !sarResult.Status.Allowed {
// 		msg := " User [ " + userId + " ] is not a owner of cluster. Cannot invtie member"
// 		klog.Infoln(msg)
// 		util.SetResponse(res, msg, nil, http.StatusBadRequest)
// 		return
// 	}

// 	// insert db
// 	if err := clusterDataFactory.Insert(clusterMember); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}

// 	// role 생성해 주면 될 듯
// 	if msg, status := caller.CreateCLMRole(clm, member, true); status != http.StatusOK {
// 		util.SetResponse(res, msg, nil, status)
// 		return
// 	}
// 	if msg, status := caller.CreateRoleInRemote(clm, member, remoteRole, true); status != http.StatusOK {
// 		util.SetResponse(res, msg, nil, status)
// 		return
// 	}

// 	// // send mail to all group member??
// 	// // body를 const에서 읽어서 replace해주고 sendMail에 body는 바로 넣자
// 	// to := []string{"sangwon_cho@tmax.co.kr"}
// 	// subject := " 신청해주신 Trial NameSpace 만료 안내 "

// 	// if err := util.SendEmail("no-reply-tc@tmax.co.kr", to, subject); err != nil {
// 	// 	klog.Errorln(err)
// 	// 	util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	//성공했다고 msg주자
// 	msg = "Group inivtation is successed"
// 	klog.Infoln(msg)
// 	util.SetResponse(res, msg, nil, http.StatusOK)
// 	return
// }

// func AcceptInvitation(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
// 	// token := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

// 	vars := gmux.Vars(req)
// 	cluster := vars["cluster"]
// 	member := vars["user"]

// 	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, member); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}

// 	clusterMember := util.ClusterMemberInfo{}
// 	clusterMember.Cluster = cluster
// 	clusterMember.Role = remoteRole
// 	clusterMember.Member = member
// 	clusterMember.Attribute = "user"
// 	clusterMember.Status = "pending"

// 	// token validation
// 	if err := util.TokenValid(req); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}
// 	// db에 pending으로 존재하는지 확인
// 	count, err := clusterDataFactory.GetPendingUser(clusterMember) //존재하는지 존재하면 status를 보고 다르게 err 넘겨줘야할 듯
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	} else if count == 0 {
// 		msg := "Invitation is expired"
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

// 	// db에 status 변경해주고 pending --> invited로..
// 	if err := clusterDataFactory.UpdateStatus(clusterMember); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}

// 	// role 생성해 주면 될 듯
// 	if msg, status := caller.CreateCLMRole(clm, member, false); status != http.StatusOK {
// 		util.SetResponse(res, msg, nil, status)
// 		return
// 	}

// 	if msg, status := caller.CreateRoleInRemote(clm, member, remoteRole, false); status != http.StatusOK {
// 		util.SetResponse(res, msg, nil, status)
// 		return
// 	}

// 	msg = "User [" + member + "] is added to cluster [" + cluster + "]"
// 	klog.Infoln(msg)
// 	util.SetResponse(res, msg, nil, status)
// 	return

// }

// func DeclineInvitation(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	// token := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

// 	vars := gmux.Vars(req)
// 	cluster := vars["cluster"]
// 	member := vars["user"]

// 	if err := util.StringParameterException(userGroups, userId, cluster, member); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}

// 	clusterMember := util.ClusterMemberInfo{}
// 	clusterMember.Cluster = cluster
// 	clusterMember.Member = member
// 	clusterMember.Attribute = "user"
// 	clusterMember.Status = "pending"

// 	// db에 status 변경해주고 pending --> invited로..
// 	if err := clusterDataFactory.Delete(clusterMember); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}
// 	msg := "User [" + member + "] reject invtiation to a cluster [" + cluster + "]"
// 	klog.Infoln(msg)
// 	util.SetResponse(res, msg, nil, http.StatusOK)
// 	return
// }

// func ListInvitation(res http.ResponseWriter, req *http.Request) {
// 	queryParams := req.URL.Query()
// 	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
// 	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
// 	cluster := queryParams.Get(QUERY_PARAMETER_CLUSTER)

// 	if err := util.StringParameterException(userGroups, userId, cluster); err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
// 		return
// 	}

// 	clusterMemberList, err := clusterDataFactory.ListPendingUser(cluster)
// 	if err != nil {
// 		klog.Errorln(err)
// 		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
// 		return
// 	}
// 	msg := "List cluster success"
// 	klog.Infoln(msg)
// 	util.SetResponse(res, msg, clusterMemberList, http.StatusOK)
// 	return
// }
