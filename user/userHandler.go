package user

import (
	"net/http"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/caller"

	rbacApi "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func Post(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** POST /user")
	queryParams := req.URL.Query()
	userId := queryParams.Get("userId")

	if userId == "" {
		out := "userId is Missing"
		util.SetResponse(res, out, nil, http.StatusBadRequest)
		klog.V(1).Infof("userId is Missing")
		return
	}

	klog.V(3).Infoln("userId is : " + userId)

	//Call CreateClusterRoleBinding for New User
	clusterRoleBinding := rbacApi.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: userId,
		},
		RoleRef: rbacApi.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "clusterrole-new-user",
		},
		Subjects: []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     userId,
			},
		},
	}
	if err := k8sApiCaller.CreateClusterRoleBinding(&clusterRoleBinding); err != nil {
		klog.V(1).Infoln(err)
		util.SetResponse(res, err.Error(), nil, http.StatusInternalServerError)
		return
	}
	out := "Create ClusterRoleBinding for New User Success"
	util.SetResponse(res, out, nil, http.StatusOK)
}

func Delete(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** DELETE /user")
	queryParams := req.URL.Query()
	userId := queryParams.Get(util.QUERY_PARAMETER_USER_ID)

	if userId == "" {
		out := "userId is Missing"
		util.SetResponse(res, out, nil, http.StatusBadRequest)
		klog.V(1).Infof("userId is Missing")
		return
	}

	klog.V(3).Infoln("userId is : " + userId)

	//Call Delete resource function for New User
	isErr := false
	if err := k8sApiCaller.DeleteClusterRoleBinding(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}
	if err := k8sApiCaller.DeleteRQCWithUser(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}
	if err := k8sApiCaller.DeleteNSCWithUser(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}
	if err := k8sApiCaller.DeleteRBCWithUser(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}
	if err := k8sApiCaller.DeleteCRBWithUser(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}
	if err := k8sApiCaller.DeleteRBWithUser(userId); err != nil {
		klog.V(1).Infoln(err)
		isErr = true
	}

	var out string
	if isErr {
		out = "Failed to completely delete related resources with " + userId
		klog.V(3).Infoln(out)
		util.SetResponse(res, out, nil, http.StatusOK)
	} else {
		out = "Successfully delete related resources with " + userId
		klog.V(3).Infoln(out)
		util.SetResponse(res, out, nil, http.StatusOK)
	}
}

func Options(res http.ResponseWriter, req *http.Request) {
	klog.V(3).Infoln("**** OPTIONS/user")
	out := "**** OPTIONS/user"
	util.SetResponse(res, out, nil, http.StatusOK)
}
