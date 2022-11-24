package cluster

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	gmux "github.com/gorilla/mux"
	util "github.com/tmax-cloud/hypercloud-api-server/util"
	caller "github.com/tmax-cloud/hypercloud-api-server/util/caller"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	"k8s.io/klog"
	// "encoding/json"
)

const (
	// QUERY_PARAMETER_USER_NAME   = "Name"
	QUERY_PARAMETER_USER_ID     = "userId"
	QUERY_PARAMETER_USER_NAME   = "userName"
	QUERY_PARAMETER_LIMIT       = "limit"
	QUERY_PARAMETER_OFFSET      = "offset"
	QUERY_PARAMETER_CLUSTER     = "cluster"
	QUERY_PARAMETER_ACCESS_ONLY = "accessOnly"
	QUERY_PARAMETER_REMOTE_ROLE = "remoteRole"
	QUERY_PARAMETER_MEMBER_NAME = "memberName"
)

func ListPage(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]
	// accessOnly := queryParams.Get(QUERY_PARAMETER_ACCESS_ONLY)
	vars := gmux.Vars(req)
	namespace := vars["namespace"]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	if namespace == "" {
		if clmList, err := caller.ListAllCluster(userId, userGroups); err != nil {
			util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clmList, http.StatusOK)
			return
		}
	} else {
		if clusterClaimList, err := caller.ListClusterInNamespace(userId, userGroups, namespace); err != nil {
			util.SetResponse(res, err.Error(), clusterClaimList, http.StatusInternalServerError)
			return
		} else {
			util.SetResponse(res, "Success", clusterClaimList, http.StatusOK)
			return
		}
	}
}

func ListLNB(res http.ResponseWriter, req *http.Request) {
	queryParams := req.URL.Query()
	userId := queryParams.Get(QUERY_PARAMETER_USER_ID)
	userGroups := queryParams[util.QUERY_PARAMETER_USER_GROUP]

	if err := util.StringParameterException(userGroups, userId); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	if clmList, err := caller.ListAccessibleCluster(userId, userGroups); err != nil {
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
	} else {
		util.SetResponse(res, "Success", clmList, http.StatusOK)
	}
}

func InsertCLM(res http.ResponseWriter, req *http.Request) {
	// queryParams := req.URL.Query()
	vars := gmux.Vars(req)
	namespace := vars["namespace"]
	clustermanager := vars["clustermanager"]

	if err := util.StringParameterException([]string{}, namespace, clustermanager); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	clm := &clusterv1alpha1.ClusterManager{}
	if err := json.Unmarshal(body, clm); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
	}

	//INSERT_QUERY, item.Namespace, item.Cluster, item.MemberId, item.MemberName, item.Attribute, item.Role, item.Status, time.Now(), time.Now())
	clmInfo := util.ClusterMemberInfo{
		Namespace:  clm.Namespace,
		Cluster:    clm.Name,
		MemberId:   clm.Annotations["owner"],
		Attribute:  "user",
		Role:       "admin",
		Status:     "owner",
		MemberName: "default",
	}

	// // 있는지 먼저 확인...
	// var res string
	// if res, err := clusterDataFactory.GetOwner(clmInfo); err != nil {
	// 	msg := "Failed to get cluster owner from db"
	// 	klog.V(3).Infoln(msg)
	// 	util.SetResponse(res, msg, nil, http.StatusInternalServerError)
	// 	return
	// }
	// // 있으니까 pass
	// if res != "" {

	// }

	if err := clusterDataFactory.Insert(clmInfo); err != nil {
		msg := "Failed to insert cluster info from db"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusInternalServerError)
		return
	}
	msg := "Success to insert cluster info from db"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}

func DeleteCLM(res http.ResponseWriter, req *http.Request) {
	// queryParams := req.URL.Query()
	vars := gmux.Vars(req)
	namespace := vars["namespace"]
	clustermanager := vars["clustermanager"]

	if err := util.StringParameterException([]string{}, namespace, clustermanager); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusBadRequest)
		return
	}

	if err := clusterDataFactory.DeleteALL(namespace, clustermanager); err != nil {
		msg := "Failed to delete cluster info from db"
		klog.V(3).Infoln(msg)
		util.SetResponse(res, msg, nil, http.StatusInternalServerError)
		return
	}
	msg := "Success to delete cluster info from db"
	klog.V(3).Infoln(msg)
	util.SetResponse(res, msg, nil, http.StatusOK)
}
