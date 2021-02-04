package caller

import (
	"context"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"

	clusterv1alpha1 "github.com/tmax-cloud/cluster-manager-operator/api/v1alpha1"
	rbacApi "k8s.io/api/rbac/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

var remoteClientset *kubernetes.Clientset
var remoteRestConfig *restclient.Config

func CreateSubjectRolebinding(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, isGroup bool) (string, int) {
	remoteClientset, err := getRemoteK8sClient(clusterManager.Name)
	if err != nil {
		return err.Error(), http.StatusForbidden
	}

	clusterRoleBindingName := subject + "-rolebinding"
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     remoteRole,
		},
	}
	if !isGroup {
		clusterRoleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		clusterRoleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}
	msg := "Create clusterrole [" + remoteRole + "] for subject [ " + subject + "] from [" + clusterManager.Name + "]"
	klog.Infoln(msg)
	return msg, http.StatusOK
}

func RemoveSubjectRolebinding(clusterManager *clusterv1alpha1.ClusterManager, subject string) (string, int) {
	remoteClientset, err := getRemoteK8sClient(clusterManager.Name)
	if err != nil {
		return err.Error(), http.StatusForbidden
	}

	clusterRoleBindingName := subject + "-rolebinding"
	_, err = remoteClientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err == nil {
		if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{}); err != nil {
			klog.Errorln(err)
			return err.Error(), http.StatusInternalServerError
		}
	} else if errors.IsNotFound(err) {
		klog.Infoln("Rolebinding [" + clusterRoleBindingName + "] is already deleted")
	} else {
		return err.Error(), http.StatusInternalServerError
	}
	msg := "Remove rolebinding [" + clusterRoleBindingName + "] for subject [ " + subject + "] from [" + clusterManager.Name + "]"
	klog.Infoln(msg)

	return msg, http.StatusOK
}

func getRemoteK8sClient(cluster string) (*kubernetes.Clientset, error) {
	remoteKubeconfig, err := Clientset.CoreV1().Secrets("capi-system").Get(context.TODO(), cluster+"-kubeconfig", metav1.GetOptions{})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}
	//
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
}
