package caller

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"

	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var remoteClientset *kubernetes.Clientset
var remoteRestConfig *restclient.Config

func CreateRoleInRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, isGroup bool) (string, int) {
	if remoteRole == "admin" {
		remoteRole = "cluster-admin"
	}
	remoteClientset, err := getRemoteK8sClient(clusterManager.Name)
	if err != nil {
		return err.Error(), http.StatusForbidden
	}

	// var clusterRoleName string
	var clusterRoleBindingName string
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if !isGroup {
		// clusterRoleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-user-rolebinding"
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		// clusterRoleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-group-rolebinding"
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	clusterRoleBinding.ObjectMeta = metav1.ObjectMeta{
		Name: clusterRoleBindingName,
	}
	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     remoteRole,
	}

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}
	msg := "Create clusterrole [" + remoteRole + "] to remote cluster [" + clusterManager.Name + "] for subject [" + subject + "] "
	klog.Infoln(msg)
	return msg, http.StatusOK
}

func RemoveRoleFromRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, isGroup bool) (string, int) {
	remoteClientset, err := getRemoteK8sClient(clusterManager.Name)
	if err != nil {
		return err.Error(), http.StatusInternalServerError
	}

	// var clusterRoleName string
	var clusterRoleBindingName string
	if !isGroup {
		clusterRoleBindingName = subject + "-user-rolebinding"
	} else {
		clusterRoleBindingName = subject + "-group-rolebinding"
	}

	_, err = remoteClientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err == nil {
		if err := remoteClientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{}); err != nil {
			klog.Errorln(err)
			return err.Error(), http.StatusInternalServerError
		}
	} else if errors.IsNotFound(err) {
		klog.Errorln("Rolebinding [" + clusterRoleBindingName + "] is already deleted")
		return err.Error(), http.StatusInternalServerError
	} else {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}
	msg := "Remove rolebinding [" + clusterRoleBindingName + "] from remote cluster [" + clusterManager.Name + "] for subject [" + subject + "]"
	klog.Infoln(msg)

	return msg, http.StatusOK
}

func getRemoteK8sClient(cluster string) (*kubernetes.Clientset, error) {
	if remoteKubeconfig, err := Clientset.CoreV1().Secrets("capi-system").Get(context.TODO(), cluster+"-kubeconfig", metav1.GetOptions{}); err == nil {
		if value, ok := remoteKubeconfig.Data["value"]; ok {
			remoteClientConfig, err := clientcmd.NewClientConfigFromBytes(value)
			if err != nil {
				klog.Errorln(err)
				return nil, err
			}
			remoteRestConfig, err = remoteClientConfig.ClientConfig()
			if err != nil {
				klog.Errorln(err)
				return nil, err
			}
		}
		remoteClientset, err = kubernetes.NewForConfig(remoteRestConfig)
		if err != nil {
			klog.Errorln(err)
			return nil, err
		}
		return remoteClientset, nil
	} else if errors.IsNotFound(err) {
		klog.Infoln("Cluster [" + cluster + "] is not ready yet")
		return nil, err
	} else {
		klog.Errorln("Error: Get clusterrole [" + cluster + "] is failed")
		return nil, err
	}
}
