package caller

import (
	"context"
	"fmt"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	coreV1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

var remoteClientset *kubernetes.Clientset
var remoteRestConfig *restclient.Config

// remote cluster에 service account와 secret을 미리 생성
func CreateSASecretInRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, attribute string) error {
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	parsed, err := util.ChangeJwtDecodeFormat(subject)
	if err != nil {
		return err
	}

	serviceAccount := &coreV1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      parsed,
			Namespace: util.KubeNamespace,
		},
	}

	_, err = remoteClientset.
		CoreV1().
		ServiceAccounts(util.KubeNamespace).
		Get(context.TODO(), serviceAccount.Name, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			ServiceAccounts(util.KubeNamespace).
			Create(context.TODO(), serviceAccount, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	serviceAccountTokenSecret := &coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				coreV1.ServiceAccountNameKey: serviceAccount.Name,
			},
			Name: serviceAccount.Name + "-token",
		},
		Type: coreV1.SecretTypeServiceAccountToken,
	}
	_, err = remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), serviceAccountTokenSecret.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := remoteClientset.
			CoreV1().
			Secrets(util.KubeNamespace).
			Create(context.TODO(), serviceAccountTokenSecret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// remote cluster에 serviceaccount를 subject로 하는 clusterrolebinding 생성
func CreateRoleInRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, attribute string) error {

	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	parsed, err := util.ChangeJwtDecodeFormat(subject)
	if err != nil {
		return err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	clusterRoleBindingName := subject + "-rolebinding"
	clusterRoleBinding.Subjects = []rbacv1.Subject{
		{
			APIGroup:  "",
			Kind:      "ServiceAccount",
			Name:      parsed,
			Namespace: util.KubeNamespace,
		},
	}

	clusterRoleBinding.ObjectMeta = metav1.ObjectMeta{
		Name: clusterRoleBindingName,
	}

	// admin인 경우 cluster-admin cluster role을 사용하므로 따로 처리
	if remoteRole == "admin" {
		remoteRole = "cluster-admin"
	}

	clusterRoleBinding.RoleRef = rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     remoteRole,
	}

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		return err
	}
	msg := "Create clusterrole [" + remoteRole + "] to remote cluster [" + clusterManager.Name + "] for subject [" + subject + "] "
	klog.V(3).Infoln(msg)
	return nil
}

// master cluster에 remote Service account secret와 동일한 secret을 생성
func CreateRemoteSecretInLocal(clusterManager *clusterv1alpha1.ClusterManager, subject string, remoteRole string, attribute string) error {
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	parsed, err := util.ChangeJwtDecodeFormat(subject)
	if err != nil {
		return err
	}

	remoteSATokenName := fmt.Sprintf("%s-%s", parsed, "token")

	remoteSASecretToken, err := remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), remoteSATokenName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		return err
	} else if err != nil {
		return err
	}

	if string(remoteSASecretToken.Data["token"]) == "" {
		return fmt.Errorf("service account token secret is not found")
	}

	localSecretName := fmt.Sprintf("%s-%s-%s", parsed, clusterManager.Name, "token")
	localTokenSecret := &coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: localSecretName,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         util.CLUSTER_API_GROUP_VERSION,
					Kind:               util.CLUSTER_API_Kind,
					Name:               clusterManager.GetName(),
					UID:                clusterManager.GetUID(),
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				},
			},
		},
		Data: map[string][]byte{
			"token": remoteSASecretToken.Data["token"],
		},
	}

	_, err = Clientset.
		CoreV1().
		Secrets(clusterManager.Namespace).
		Get(context.TODO(), localSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err := Clientset.
			CoreV1().
			Secrets(clusterManager.Namespace).
			Create(context.TODO(), localTokenSecret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// remote cluster에 생성한 clusterrolebinding 삭제
func RemoveRoleFromRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	clusterRoleBindingName := subject + "-rolebinding"

	if _, err := remoteClientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infoln("Rolebinding [" + clusterRoleBindingName + "] is already deleted")
			return nil
		} else {
			return err
		}
	} else {
		if err := remoteClientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	msg := "Remove rolebinding [" + clusterRoleBindingName + "] from remote cluster [" + clusterManager.Name + "] for subject [" + subject + "]"
	klog.V(3).Infoln(msg)
	return nil
}

// remote cluster에 생성한 service account, service account secret 모두 삭제
func RemoveSASecretFromRemote(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	remoteClientset, err := getRemoteK8sClient(clusterManager)
	if err != nil {
		return err
	}

	parsed, err := util.ChangeJwtDecodeFormat(subject)
	if err != nil {
		return err
	}

	_, err = remoteClientset.
		CoreV1().
		Secrets(util.KubeNamespace).
		Get(context.TODO(), parsed+"-token", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		klog.V(3).Infoln("Secret [" + parsed + "-token] is already deleted")
		return nil
	} else if err != nil {
		return err
	}
	if err := remoteClientset.CoreV1().Secrets(util.KubeNamespace).Delete(context.TODO(), parsed+"-token", metav1.DeleteOptions{}); err != nil {
		return err
	}

	_, err = remoteClientset.
		CoreV1().
		ServiceAccounts(util.KubeNamespace).
		Get(context.TODO(), parsed, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		klog.V(3).Infoln("ServiceAccount [" + parsed + "] is already deleted")
		return nil
	} else if err != nil {
		return err
	}

	if err := remoteClientset.CoreV1().ServiceAccounts(util.KubeNamespace).Delete(context.TODO(), parsed, metav1.DeleteOptions{}); err != nil {
		return err
	}

	msg := "Remove serviceaccount and secret [" + parsed + "] from remote cluster [" + clusterManager.Name + "] for subject [" + subject + "]"
	klog.V(3).Infoln(msg)
	return nil
}

// master cluster에 생성한 remote cluster의 service account secret token을 삭제
func RemoveRemoteSecretInLocal(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {

	parsed, err := util.ChangeJwtDecodeFormat(subject)
	if err != nil {
		return err
	}

	localSecretName := fmt.Sprintf("%s-%s-%s", parsed, clusterManager.Name, "token")

	_, err = Clientset.
		CoreV1().
		Secrets(clusterManager.Namespace).
		Get(context.TODO(), localSecretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		klog.V(3).Infoln("Secret [" + localSecretName + "] is already deleted")
		return nil
	} else if err != nil {
		return err
	}

	err = Clientset.
		CoreV1().
		Secrets(clusterManager.Namespace).
		Delete(context.TODO(), localSecretName, metav1.DeleteOptions{})
	if err != nil {
		klog.V(3).Infoln(err)
		return err
	}

	msg := "Remove secret token [" + localSecretName + "] from local cluster for subject [" + subject + "]"
	klog.V(3).Infoln(msg)
	return nil
}

// CAPI가 생성한 kubeconfig를 통해서 remote cluster에 접근하기 위한 clientset을 반환
func getRemoteK8sClient(clusterManager *clusterv1alpha1.ClusterManager) (*kubernetes.Clientset, error) {
	if remoteKubeconfig, err := Clientset.CoreV1().Secrets(clusterManager.Namespace).Get(context.TODO(), clusterManager.Name+"-kubeconfig", metav1.GetOptions{}); err == nil {
		if value, ok := remoteKubeconfig.Data["value"]; ok {
			remoteClientConfig, err := clientcmd.NewClientConfigFromBytes(value)
			if err != nil {
				return nil, err
			}
			remoteRestConfig, err = remoteClientConfig.ClientConfig()
			if err != nil {
				return nil, err
			}
		}
		remoteClientset, err = kubernetes.NewForConfig(remoteRestConfig)
		if err != nil {
			return nil, err
		}
		return remoteClientset, nil
	} else if errors.IsNotFound(err) {
		klog.V(3).Infoln("Cluster [" + clusterManager.Name + "] is not ready yet")
		return nil, err
	} else {
		klog.V(1).Infoln("Error: Get clusterrole [" + clusterManager.Name + "] is failed")
		return nil, err
	}
}
