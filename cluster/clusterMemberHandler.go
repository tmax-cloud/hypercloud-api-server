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

	clm, msg, status := caller.GetCluster(userId, userGroups, cluster)
	if clm == nil {
		util.SetResponse(res, msg, nil, status)
		return
	}

	// var clusterOwner string
	// var test []util.ClusterMemberInfo
	// for _, val := range clusterMemberList {
	// 	if val.Status == "owner" {
	// 		clusterOwner = val.MemberId
	// 		test = append(test, val)
	// 	} else if val.Status != "pending" {
	// 		test = append(test, val)
	// 	}
	// }

	if userId == clm.Annotations["owner"] {
		clusterMemberList, err := clusterDataFactory.ListAllClusterMember(cluster)
		if err != nil {
			klog.Errorln(err)
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		}
		msg = "List cluster success"
		klog.Infoln(msg)
		util.SetResponse(res, msg, clusterMemberList, http.StatusOK)
		return
	} else {
		clusterMemberList, err := clusterDataFactory.ListClusterMemberWithOutPending(cluster)
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
