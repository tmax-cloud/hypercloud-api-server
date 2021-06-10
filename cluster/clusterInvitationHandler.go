package cluster

// "encoding/json"
import (
	"net/http"
	"strings"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"

	// caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	"k8s.io/klog"
	// "encoding/json"
)

const (
	LINK = "https://@@CONSOLE_LB@@/api/hypercloud/namespaces/@@NAMESPACE@@/clustermanagers/@@CLUSTER_NAME@@/member_invitation/user/@@MEMBER_ID@@/accept?userId=@@OWNER_EMAIL@@&token=@@TOKEN@@"
)

func InviteUser(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// userName := queryParams.Get(QUERY_PARAMETER_USER_NAME)
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)
	memberName := queryParams.Get(QUERY_PARAMETER_MEMBER_NAME)
	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	memberId := vars["member"]
	clusterManagerNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId, clusterManagerNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = clusterManagerNamespace
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.MemberName = memberName
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// cluster ready 인지 확인
	clm, msg, status := caller.GetCluster(userId, userGroups, cluster, clusterManagerNamespace)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if clm.Status.Ready == false || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(clusterMember.Cluster, clusterMember.Namespace)
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

	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", clusterManagerNamespace, cluster, "update")
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

	var b strings.Builder
	b.WriteString(LINK)
	for _, userGroup := range userGroups {
		b.WriteString("?userGroup=")
		b.WriteString(userGroup)
	}

	to := []string{memberId}
	from := "no-reply-tc@tmax.co.kr"
	subject := userId + "(이)가 당신을 " + cluster + " cluster에 초대하였습니다."
	bodyParameter := map[string]string{}
	// bodyParameter["@@LINK@@"] = LINK
	bodyParameter["@@LINK@@"] = "https://" + ConsoleLB + "/k8s/all-namespaces/clustermanagers"
	bodyParameter["@@CLUSTER_NAME@@"] = cluster
	bodyParameter["@@VALID_TIME@@"] = util.ValidTime
	bodyParameter["@@OWNER_EMAIL@@"] = userId
	bodyParameter["@@OWNER_NAME@@"] = userId
	bodyParameter["@@CONSOLE_LB@@"] = ConsoleLB
	bodyParameter["@@NAMESPACE@@"] = clusterMember.Namespace
	bodyParameter["@@MEMBER_ID@@"] = clusterMember.MemberId
	bodyParameter["@@TOKEN@@"] = token

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
	cluster := vars["clustermanager"]
	memberId := vars["member"]
	clusterManagerNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId, clusterManagerNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = clusterManagerNamespace
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "group"
	clusterMember.Status = "invited"

	// cluster ready 인지 확인
	clm, msg, status := caller.GetCluster(userId, userGroups, cluster, clusterManagerNamespace)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if clm.Status.Ready == false || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterGroup(clusterMember.Cluster, clusterMember.Namespace)
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
			//existGroup = append(existGroup, val.MemberId)
		} else {
			existGroup = append(existGroup, val.MemberId)
		}
	}

	if userId != clusterOwner {
		msg := "Request user [ " + userId + " ]is not a cluster owner [ " + clusterOwner + " ]"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if util.Contains(existGroup, memberId) {
		msg := "Group [ " + memberId + " ] is already invited in cluster [ " + cluster + " ] "
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", clusterManagerNamespace, cluster, "update")
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
	// 1. NS get 주고..
	if msg, status := caller.CreateNSGetRole(clm, memberId, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

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
	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	memberId := vars["member"]
	clusterManagerNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, memberId, clusterManagerNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	consoleService, err := caller.GetConsoleService("console-system", "console")
	ConsoleLB := consoleService.Status.LoadBalancer.Ingress[0].IP
	if err != nil {
		klog.Infoln(err)
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = clusterManagerNamespace
	clusterMember.Cluster = cluster
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// token validation
	if _, err := util.TokenValid(req, clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	/////////////////////////////////////

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(clusterMember.Cluster, clusterMember.Namespace)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// 초대할 권한이 있는지 확인
	var clusterOwner string
	var pendingUser string
	var pendingUserRole string
	for _, val := range clusterMemberList {
		if val.Status == "owner" {
			clusterOwner = val.MemberId
			//existGroup = append(existGroup, val.MemberId)
		} else if val.Status == "pending" && val.MemberId == memberId {
			pendingUser = val.MemberId
			pendingUserRole = val.Role
		} else if val.Status == "invited" && val.MemberId == memberId {
			http.Redirect(res, req, "https://"+ConsoleLB, http.StatusSeeOther)
			msg := "User [" + memberId + "] is already invtied to cluster [" + cluster + "]"
			klog.Infoln(msg)
			util.SetResponse(res, msg, nil, http.StatusOK)
			return
		}
	}

	if userId != clusterOwner {
		msg := "Request user [ " + userId + " ]is not a cluster owner [ " + clusterOwner + " ]"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if pendingUser == "" {
		msg := "The invitation for User [ " + memberId + " ] is expired in cluster [ " + cluster + " ] "
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMember.Role = pendingUserRole

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster, clusterManagerNamespace)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if clm.Status.Ready == false || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
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
	if msg, status := caller.CreateNSGetRole(clm, memberId, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if msg, status := caller.CreateCLMRole(clm, memberId, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	if msg, status := caller.CreateRoleInRemote(clm, memberId, clusterMember.Role, clusterMember.Attribute); status != http.StatusOK {
		util.SetResponse(res, msg, nil, status)
		return
	}

	/// redirection
	http.Redirect(res, req, "https://"+ConsoleLB, http.StatusSeeOther)

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
	cluster := vars["clustermanager"]
	memberId := vars["member"]
	clusterManagerNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, memberId, clusterManagerNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = clusterManagerNamespace
	clusterMember.Cluster = cluster
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// token validation
	if _, err := util.TokenValid(req, clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(clusterMember.Cluster, clusterMember.Namespace)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// 거절 할 권한이 있는지 확인
	var clusterOwner string
	var pendingUser string
	for _, val := range clusterMemberList {
		if val.Status == "owner" {
			clusterOwner = val.MemberId
		} else if val.Status == "pending" && val.MemberId == memberId {
			pendingUser = val.MemberId
		}
	}

	if userId != clusterOwner {
		msg := "Request user [ " + userId + " ]is not a cluster owner [ " + clusterOwner + " ]"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if pendingUser == "" {
		msg := "The invitation for User [ " + memberId + " ] is expired in cluster [ " + cluster + " ] "
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if err := clusterDataFactory.Delete(clusterMember); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	msg := "Invtiation for User [" + memberId + "] is rejected in a cluster [" + cluster + "]"
	klog.Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
	return
}

func ListInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	clusterManagerNamespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, clusterManagerNamespace); err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster, clusterManagerNamespace)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}
	if clm.Status.Ready == false || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(cluster, clusterManagerNamespace)
	if err != nil {
		klog.Errorln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	var clusterOwner string
	var pendingUser []util.ClusterMemberInfo
	for _, val := range clusterMemberList {
		if val.Status == "owner" {
			clusterOwner = val.MemberId
		} else if val.Status == "pending" {
			pendingUser = append(pendingUser, val)
		}
	}

	if userId != clusterOwner {
		msg := "Request user [ " + userId + " ]is not a cluster owner [ " + clusterOwner + " ]"
		klog.Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	msg = "List cluster success"
	klog.Infoln(msg)
	util.SetResponse(res, msg, pendingUser, http.StatusOK)
	return
}
