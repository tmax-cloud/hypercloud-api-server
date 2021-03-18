package cluster

// "encoding/json"
import (
	"net/http"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"

	// caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	"k8s.io/klog"
	// "encoding/json"
)

func InviteUser(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	userName := queryParams.Get(QUERY_PARAMETER_USER_NAME)
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
	memberName := queryParams.Get(QUERY_PARAMETER_MEMBER_NAME)
	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	memberId := vars["member"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.MemberName = memberName
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// cluster ready 인지 확인
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

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(clusterMember.Cluster)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// 초대할 권한이 있는지 확인
	var clusterOwner string
	var existUser []string
	for _, val := range clusterMemberList {
		if val.Status == "owner" {
			clusterOwner = val.MemberId
			existUser = append(existUser, val.MemberId)
		} else {
			existUser = append(existUser, val.MemberId)
		}
	}
	if userId != clusterOwner {
		msg := "Request user is not a cluster owner"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if util.Contains(existUser, memberId) {
		msg := "Member is already invited in cluster"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "update")
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !sarResult.Status.Allowed {
		msg := " User [ " + userId + " ] is not a owner of cluster or owner role is deleted."
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// create Token
	token, err := util.CreateToken(clusterMember)
	klog.Info("token = \n" + token)

	// insert db
	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	consoleService, err := caller.GetConsoleService("console-system", "console")
	ConsoleLB := consoleService.Status.LoadBalancer.Ingress[0].IP
	if err != nil {
		panic(err)
	}

	to := []string{memberId}
	from := "no-reply-tc@tmax.co.kr"
	subject := userId + "(이)가 당신을 " + cluster + " kubernetes cluster에 초대하였습니다."
	bodyParameter := map[string]string{}
	bodyParameter["%%TO%%"] = to[0]
	bodyParameter["%%FROM%%"] = userId
	bodyParameter["%%CLUSTER%%"] = cluster
	bodyParameter["%%ATTRIBUTE%%"] = clusterMember.Attribute
	bodyParameter["%%USER%%"] = clusterMember.MemberId
	bodyParameter["%%IP%%"] = ConsoleLB
	bodyParameter["%%TOKEN%%"] = token
	bodyParameter["%%DAY%%"] = util.TokenExpiredDate
	bodyParameter["%%USER_NAME%%"] = userName
	bodyParameter["%%MEMBER_NAME%%"] = memberName

	if err := util.SendEmail(from, to, subject, bodyParameter); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		if err := clusterDataFactory.Delete(clusterMember); err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		}
		return
	}

	msg = "User inivtation is successed"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
	return
}

func InviteGroup(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	memberId := vars["member"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "group"
	clusterMember.Status = "invited"

	// cluster ready 인지 확인
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

	/////////////////////
	clusterMemberList, err := clusterDataFactory.ListAllClusterGroup(clusterMember.Cluster)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// 초대할 권한이 있는지 확인
	var clusterOwner string
	var existGroup []string
	for _, val := range clusterMemberList {
		if val.Status == "owner" {
			clusterOwner = val.MemberId
			existGroup = append(existGroup, val.MemberId)
		} else {
			existGroup = append(existGroup, val.MemberId)
		}
	}

	if userId != clusterOwner {
		msg := "Request user is not a cluster owner"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if util.Contains(existGroup, memberId) {
		msg := "Member is already invited in cluster"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// // db에 이미 있는지 확인
	// count, err := clusterDataFactory.GetInvitedGroup(clusterMember)
	// if err != nil {
	// 	klog.Errorln(err)
	// 	util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
	// 	return
	// } else if count != 0 {
	// 	msg := "Group is already member of cluster"
	// 	klog.Infoln(msg)
	// 	util.SetResponse(res, msg, nil, http.StatusBadRequest)
	// 	return
	// }

	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "update")
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !sarResult.Status.Allowed {
		msg := " User [ " + userId + " ] is not a owner of cluster. Cannot invtie member"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// insert db
	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// role 생성해 주면 될 듯
	if msg, status := caller.CreateCLMRole(clm, memberId, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if msg, status := caller.CreateRoleInRemote(clm, memberId, remoteRole, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	msg = "Group inivtation is successed"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
	return
}

func AcceptInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
	// token := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
	// memberName := queryParams.Get(QUERY_PARAMETER_USER_Name)
	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	memberId := vars["user"]

	if err := util.StringParameterException(userGroups, userId, cluster, memberId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	// clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// token validation
	if err := util.TokenValid(req, clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}
	// db에 pending으로 존재하는지 확인
	clusterMemberList, err := clusterDataFactory.GetPendingUser(clusterMember) //존재하는지 존재하면 status를 보고 다르게 err 넘겨줘야할 듯
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	} else if len(clusterMemberList) == 0 {
		msg := "Invitation is expired"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMember.Role = clusterMemberList[0].Role

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

	// db에 status 변경해주고 pending --> invited로..
	if err := clusterDataFactory.UpdateStatus(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// role 생성해 주면 될 듯
	if msg, status := caller.CreateCLMRole(clm, memberId, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if msg, status := caller.CreateRoleInRemote(clm, memberId, clusterMember.Role, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	msg = "User [" + memberId + "] is added to cluster [" + cluster + "]"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, status)
	return

}

func DeclineInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// token := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	vars := gmux.Vars(req)
	cluster := vars["cluster"]
	memberId := vars["user"]

	if err := util.StringParameterException(userGroups, userId, cluster, memberId); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Cluster = cluster
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// db에 status 변경해주고 pending --> invited로..
	if err := clusterDataFactory.Delete(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	msg := "User [" + memberId + "] reject invtiation to a cluster [" + cluster + "]"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
	return
}

func ListInvitation(res http.ResponseWriter, req *http.Request) {
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

	if clm, msg, status := caller.GetCluster(userId, userGroups, cluster); clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListPendingUser(cluster)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	msg := "List cluster success"
	klog.Infoln(msg)
	util.SetResponse(res, msg, clusterMemberList, http.StatusOK)
	return
}
