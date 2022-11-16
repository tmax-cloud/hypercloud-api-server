package cluster

// "encoding/json"
import (
	"fmt"
	"net/http"
	"os"
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
	// LINK = "https://@@CONSOLE_LB@@/api/hypercloud/namespaces/@@NAMESPACE@@/clustermanagers/@@CLUSTER_NAME@@/member_invitation/accept?userId=@@MEMBER_EMAIL@@&token=@@TOKEN@@"
	LINK = "https://@@CONSOLE_LB@@/k8s/ns/@@NAMESPACE@@/clustermanagers/@@CLUSTER_NAME@@/access/accept"
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
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId, namespace); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = namespace
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.MemberName = memberName
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// cluster ready 인지 확인
	clm, err := caller.GetCluster(userId, userGroups, cluster, namespace)
	if err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	if !clm.Status.Ready || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(clusterMember.Cluster, clusterMember.Namespace)
	if err != nil {
		klog.V(1).Infoln(err)
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
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if util.Contains(existUser, memberId) {
		msg := "Member is already invited in cluster"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	sarResult, err := caller.CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", namespace, cluster, "update")
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !sarResult.Status.Allowed {
		msg := " User [ " + userId + " ] is not a owner of cluster or owner role is deleted."
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// create Token
	// token, err := util.CreateToken(clusterMember)
	// klog.V(3).Info("token = \n" + token)

	// insert db
	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// consoleService, err := caller.GetConsoleService("console-system", "console")
	// ConsoleLB := consoleService.Status.LoadBalancer.Ingress[0].IP
	// if err != nil {
	// 	panic(err)
	// }

	//consoleIngress, err := caller.GetConsoleIngress("api-gateway-system", "console")
	ConsoleDomain := os.Getenv("CONSOLE_SUBDOMAIN") + "." + os.Getenv("HC_DOMAIN")

	var b strings.Builder
	b.WriteString(LINK)
	// for _, userGroup := range userGroups {
	// 	b.WriteString("&userGroup=")
	// 	b.WriteString(userGroup)
	// }

	to := []string{memberId}
	from := "no-reply-tc@tmax.co.kr"
	subject := userId + "(이)가 당신을 " + cluster + " cluster에 초대하였습니다."
	bodyParameter := map[string]string{}
	bodyParameter["@@LINK@@"] = b.String()
	fmt.Println("#########################" + bodyParameter["@@LINK@@"])
	bodyParameter["@@LINK@@"] = "https://" + ConsoleDomain + "/k8s/ns/" + namespace + "/clustermanagers/" + cluster + "/access/accept"
	bodyParameter["@@CLUSTER_NAME@@"] = cluster
	bodyParameter["@@VALID_TIME@@"] = util.ValidTime
	bodyParameter["@@OWNER_EMAIL@@"] = userId
	bodyParameter["@@MEMBER_EMAIL@@"] = memberId
	bodyParameter["@@OWNER_NAME@@"] = userId
	bodyParameter["@@CONSOLE_LB@@"] = ConsoleDomain
	bodyParameter["@@NAMESPACE@@"] = clusterMember.Namespace
	bodyParameter["@@MEMBER_ID@@"] = clusterMember.MemberId
	// bodyParameter["@@TOKEN@@"] = token
	bodyParameter["@@ROLE@@"] = remoteRole

	if err := util.SendEmail(from, to, subject, bodyParameter); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		if err := clusterDataFactory.Delete(clusterMember); err != nil {
			klog.V(1).Infoln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		}
		return
	}

	msg := "User is invited successfully"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}

func InviteGroup(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	remoteRole := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	memberId := vars["member"]
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, remoteRole, cluster, memberId, namespace); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = namespace
	clusterMember.Cluster = cluster
	clusterMember.Role = remoteRole
	clusterMember.MemberId = memberId
	clusterMember.Attribute = "group"
	clusterMember.Status = "invited"

	// cluster ready 인지 확인
	clm, err := caller.GetCluster(userId, userGroups, cluster, namespace)
	if err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	if !clm.Status.Ready || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// 클러스터에 속한 group들과 소유자(owner)반환
	clusterMemberList, err := clusterDataFactory.ListClusterOwnerAndGroupMember(clusterMember.Cluster, clusterMember.Namespace)
	if err != nil {
		klog.V(1).Infoln(err)
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
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	if util.Contains(existGroup, memberId) {
		msg := "Group [ " + memberId + " ] is already invited in cluster [ " + cluster + " ] "
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// insert db
	if err := clusterDataFactory.Insert(clusterMember); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if err := caller.CreateNSGetRole(clm, clusterMember.MemberId, clusterMember.Attribute); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if err := caller.CreateCLMRole(clm, clusterMember.MemberId, clusterMember.Attribute); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if err := caller.CreateRoleInRemote(clm, clusterMember.MemberId, remoteRole, clusterMember.Attribute); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	msg := "Invite group to cluster successfully"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}

func AcceptInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, namespace); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	// consoleService, err := caller.GetConsoleService("console-system", "console")
	// ConsoleLB := consoleService.Status.LoadBalancer.Ingress[0].IP
	// if err != nil {
	// 	klog.V(3).Infoln(err)
	// }

	ConsoleDomain := os.Getenv("CONSOLE_SUBDOMAIN") + "." + os.Getenv("HC_DOMAIN")

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = namespace
	clusterMember.Cluster = cluster
	clusterMember.MemberId = userId
	clusterMember.Attribute = "user"

	// token validation
	// if _, err := util.TokenValid(req, clusterMember); err != nil {
	// 	klog.V(1).Infoln(err)
	// 	util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
	// 	return
	// }
	// 해당 클러스터에 사용자 있는지 조회
	// res 없다면,,,  거절당한거 (timeout인거는 token에서 거를꺼고.. )
	// 있는데 상태가 invited면 이미 있네
	// 있는데 상태가 pending이면 나머지 로직 수행

	pendingUser, err := clusterDataFactory.GetPendingUser(clusterMember)
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	// clusterMember.Role = pendingUser.Role
	// switch pendingUser.Status {
	// case "":

	// }

	if pendingUser.Status == "" {
		msg := "Invitation for user [" + userId + "] to cluster [" + cluster + "] is not exist"
		klog.V(3).Info(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else if pendingUser.Status == "invited" {
		http.Redirect(res, req, "https://"+ConsoleDomain, http.StatusSeeOther)
		msg := "User [" + userId + "] is already invtied to cluster [" + cluster + "]"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusOK)
		return
	}

	// var clm *clusterv1alpha1.ClusterManager
	clm, err := caller.GetClusterWithoutSAR(userId, userGroups, cluster, namespace)
	if err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if !clm.Status.Ready || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	// db에 status 변경해주고 pending --> invited로..
	if err := clusterDataFactory.UpdateStatus(pendingUser); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// role 생성해 주면 될 듯
	if err := caller.CreateNSGetRole(clm, userId, pendingUser.Attribute); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	if err := caller.CreateCLMRole(clm, userId, pendingUser.Attribute); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// remote cluster에 service account 및 secret 미리 생성
	if err := caller.CreateSASecretInRemote(clm, userId, pendingUser.Role, clusterMember.Attribute); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// remote service account를 subject로 하는 cluster role binding 생성
	if err := caller.CreateRoleInRemote(clm, userId, pendingUser.Role, clusterMember.Attribute); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	// remote service account secret을 가져와 master cluster에 service account 및 secret 생성
	if err := caller.CreateRemoteSecretInLocal(clm, userId, pendingUser.Role, clusterMember.Attribute); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	/// redirection
	http.Redirect(res, req, "https://"+ConsoleDomain+"/k8s/ns/"+namespace+"/clustermanagers", http.StatusSeeOther)

	msg := "User [" + userId + "] is added to cluster [" + cluster + "]"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}

func DeclineInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// token := queryParams.Get(QUERY_PARAMETER_REMOTE_ROLE)

	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, userId, namespace); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clusterMember := util.ClusterMemberInfo{}
	clusterMember.Namespace = namespace
	clusterMember.Cluster = cluster
	clusterMember.MemberId = userId
	clusterMember.Attribute = "user"
	clusterMember.Status = "pending"

	// token validation
	// if _, err := util.TokenValid(req, clusterMember); err != nil {
	// 	klog.V(1).Infoln(err)
	// 	util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
	// 	return
	// }

	pendingUser, err := clusterDataFactory.GetPendingUser(clusterMember)
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}

	ConsoleDomain := os.Getenv("CONSOLE_SUBDOMAIN") + "." + os.Getenv("HC_DOMAIN")
	// consoleService, err := caller.GetConsoleService("console-system", "console")
	// ConsoleLB := consoleService.Status.LoadBalancer.Ingress[0].IP
	// if err != nil {
	// 	klog.V(3).Infoln(err)
	// }
	if pendingUser.Status == "" {
		msg := "Invitation for user [" + userId + "] is expired to cluster [" + cluster + "]"
		klog.V(3).Info(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	} else if pendingUser.Status == "invited" {
		http.Redirect(res, req, "https://"+ConsoleDomain, http.StatusSeeOther)
		msg := "User [" + userId + "] is already invtied to cluster [" + cluster + "]"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusOK)
		return
	}

	if err := clusterDataFactory.Delete(clusterMember); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	msg := "Invitation for User [" + userId + "] is rejected in a cluster [" + cluster + "]"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}

func ListInvitation(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	vars := gmux.Vars(req)
	cluster := vars["clustermanager"]
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId, cluster, namespace); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	// cluster ready 인지 확인
	clm, err := caller.GetCluster(userId, userGroups, cluster, namespace)
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	if !clm.Status.Ready || clm.Status.Phase == "Deleting" {
		msg := "Cannot invite member to cluster in deleting phase or not ready status"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	clusterMemberList, err := clusterDataFactory.ListAllClusterUser(cluster, namespace)
	if err != nil {
		klog.V(1).Infoln(err)
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
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusBadRequest)
		return
	}

	msg := "List cluster success"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, pendingUser, http.StatusOK)
}
