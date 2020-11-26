package Caller

import (
	"bytes"
	"context"
	"encoding/json"
	"hypercloud-api-server/util"
	"io"
	"reflect"
	"sync"

	claim "github.com/tmax-cloud/hypercloud-go-operator/api/v1alpha1"

	authApi "k8s.io/api/authorization/v1"
	coreApi "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacApi "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

var Clientset *kubernetes.Clientset
var config *restclient.Config

func init() {

	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "/root/.kube")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "/root/.kube")
	// }
	// flag.Parse()

	// var err error
	// config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	klog.Errorln(err)
	// 	panic(err)
	// }
	// config.Burst = 100
	// config.QPS = 100
	// Clientset, err = kubernetes.NewForConfig(config)
	// if err != nil {
	// 	klog.Errorln(err)
	// 	panic(err)
	// }

	// If api-server on POD, activate below code and delete above
	// creates the in-cluster config
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	config.Burst = 100
	config.QPS = 100
	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

}

func CreateClusterRoleBinding(ClusterRoleBinding *rbacApi.ClusterRoleBinding) {
	result, err := Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), ClusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	klog.Info(" Create ClusterRoleBinding " + result.GetObjectMeta().GetName() + " Success ")
}

func DeleteClusterRoleBinding(name string) {
	deletePolicy := metav1.DeletePropagationForeground
	if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		klog.Errorln(err)
		panic(err)
	}
	klog.Info(" Delete ClusterRoleBinding " + name + " Success ")
}

func GetAccessibleNS(userId string, labelSelector string) coreApi.NamespaceList {
	var nsList = &coreApi.NamespaceList{}
	// 1. Get UserGroup List if Exists
	klog.Infoln("userId : ", userId)
	userDetail := getUserDetailWithoutToken(userId)
	var userGroups []string
	if userDetail["groups"] != nil {
		for _, userGroup := range userDetail["groups"].([]interface{}) {
			userGroups = append(userGroups, userGroup.(string))
		}
	}
	for _, userGroup := range userGroups {
		klog.Infoln("userGroupName : ", userGroup)
	}

	// 2. Check If User has NS List Role
	nsListRuleReview := authApi.SubjectAccessReview{
		Spec: authApi.SubjectAccessReviewSpec{
			ResourceAttributes: &authApi.ResourceAttributes{
				Resource: "namespaces",
				Verb:     "list",
				Group:    "",
			},
			User:   userId,
			Groups: userGroups,
		},
	}
	sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nsListRuleReview, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	if sarResult.Status.Allowed {
		klog.Infoln(" User [ ", userId, " ] has Namespace List Role, Can Access All Namespace")
		nsList, err = Clientset.CoreV1().Namespaces().List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: labelSelector,
			},
		)
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
	} else {
		klog.Infoln(" User [ ", userId, " ] has No Namespace List Role, Check If user has Namespace Get Role to Certain Namespace")
		potentialNsList, err := Clientset.CoreV1().Namespaces().List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: labelSelector,
			},
		)
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
		var wg sync.WaitGroup
		wg.Add(len(potentialNsList.Items))
		for _, potentialNs := range potentialNsList.Items {
			go func(potentialNs coreApi.Namespace, userId string, userGroups []string, nsList *coreApi.NamespaceList) {
				defer wg.Done()
				nsGetRuleReview := authApi.SubjectAccessReview{
					Spec: authApi.SubjectAccessReviewSpec{
						ResourceAttributes: &authApi.ResourceAttributes{
							Resource:  "namespaces",
							Verb:      "get", //FIXME : list??
							Group:     "",
							Namespace: potentialNs.GetName(),
						},
						User:   userId,
						Groups: userGroups,
					},
				}
				sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nsGetRuleReview, metav1.CreateOptions{})
				if err != nil {
					klog.Errorln(err)
					panic(err)
				}
				if sarResult.Status.Allowed {
					klog.Infoln(" User [ ", userId, " ] has Namespace Get Role in Namspace [ ", potentialNs.GetName(), " ]")
					nsList.Items = append(nsList.Items, potentialNs)
				}
			}(potentialNs, userId, userGroups, nsList)
		}
		wg.Wait()

		if len(nsList.Items) > 0 {
			nsList.APIVersion = potentialNsList.APIVersion
			nsList.Continue = potentialNsList.Continue
			nsList.ResourceVersion = potentialNsList.ResourceVersion
			nsList.TypeMeta = potentialNsList.TypeMeta
		} else {
			klog.Infoln(" User [ ", userId, " ] has No Namespace Get Role in Any Namspace")
		}
	}
	if len(nsList.Items) > 0 {
		klog.Infoln("=== [ ", userId, " ] Accessible Namespace ===")
		for _, ns := range nsList.Items {
			klog.Infoln("  ", ns.Name)
		}
	}
	return *nsList
}

func GetAccessibleNSC(userId string, labelSelector string) claim.NamespaceClaimList {
	// var nsList = &coreApi.NamespaceList{}
	var nscList = &claim.NamespaceClaimList{}

	// 1. Check If User has NSC List Role
	nsListRuleReview := authApi.SubjectAccessReview{
		Spec: authApi.SubjectAccessReviewSpec{
			ResourceAttributes: &authApi.ResourceAttributes{
				Resource: "namespaceclaims",
				Verb:     "list",
				Group:    util.HYPERCLOUD4_CLAIM_API_GROUP,
			},
			User: userId,
		},
	}

	sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nsListRuleReview, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}
	klog.Infoln("sarResult : " + sarResult.String())

	// /apis/claim.tmax.io/v1alpha1/namespaceclaims?labelselector
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").Param(util.QUERY_PARAMETER_LABEL_SELECTOR, labelSelector).DoRaw(context.TODO())
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}

	if sarResult.Status.Allowed {
		klog.Infoln(" User [ ", userId, " ] has NamespaceClaim List Role, Can Access All NamespaceClaim")

		if err := json.Unmarshal(data, &nscList); err != nil {
			klog.Errorln(err)
			panic(err)
		}

	} else {
		klog.Infoln(" User [ ", userId, " ] has No NamespaceClaim List Role, Check If user has NamespaceClaim Get Role & has Owner Annotation on certain NamespaceClaim")
		// 2. Check If User has NSC Get Role
		nscGetRuleReview := authApi.SubjectAccessReview{
			Spec: authApi.SubjectAccessReviewSpec{
				ResourceAttributes: &authApi.ResourceAttributes{
					Resource: "namespaceclaims",
					Verb:     "get",
					Group:    util.HYPERCLOUD4_CLAIM_API_GROUP,
				},
				User: userId,
			},
		}

		sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nscGetRuleReview, metav1.CreateOptions{})
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
		if sarResult.Status.Allowed {
			klog.Infoln(" User [ ", userId, " ] has NamespaceClaim Get Role")
			var potentialNscList = &claim.NamespaceClaimList{}
			if err := json.Unmarshal(data, &potentialNscList); err != nil {
				klog.Errorln(err)
				panic(err)
			}

			var wg sync.WaitGroup
			wg.Add(len(potentialNscList.Items))
			for _, potentialNsc := range potentialNscList.Items {
				go func(potentialNsc claim.NamespaceClaim, userId string, nscList *claim.NamespaceClaimList) {
					defer wg.Done()
					if potentialNsc.Annotations["owner"] == userId {
						klog.Infoln(" User [ ", userId, " ] has owner annotation in NamspaceClaim [ ", potentialNsc.Name, " ]")
						nscList.Items = append(nscList.Items, potentialNsc)
					}
				}(potentialNsc, userId, nscList)
			}
			wg.Wait()

			if len(nscList.Items) > 0 {
				nscList.APIVersion = potentialNscList.APIVersion
				nscList.Continue = potentialNscList.Continue
				nscList.ResourceVersion = potentialNscList.ResourceVersion
				nscList.TypeMeta = potentialNscList.TypeMeta
			} else {
				klog.Infoln(" User [ ", userId, " ] has No owner annotaion in Any NamspaceClaim")
			}
		} else {
			klog.Infoln(" User [ ", userId, " ] has no NamespaceClaim Get Role, User Cannot Access any NamespaceClaim")
		}

	}

	if len(nscList.Items) > 0 {
		klog.Infoln("=== [ ", userId, " ] Accessible NamespaceClaim ===")
		for _, nsc := range nscList.Items {
			klog.Infoln("  ", nsc.Name)
		}
	}
	return *nscList
}

// ExecCommand sends a 'exec' command to specific pod.
// It returns outputs of command.
// If the container parameter == "", it chooses first container.
func ExecCommand(pod v1.Pod, command []string, container string) (string, string, error) {

	var stdin io.Reader

	req := Clientset.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).
		Namespace(pod.Namespace).SubResource("exec")

	if container == "" {
		container = pod.Spec.Containers[0].Name
	}

	option := &v1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}
	if stdin == nil {
		option.Stdin = false
	}

	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}
	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	if err != nil {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

// GetPodListByLabel returns a pod List using label and namespace.
// If you want to find pods through all namespace, pass "" for namespace parameter.
// If there is a pod list, it returns a list with 'true', if not, returns with 'false'
func GetPodListByLabel(label string, namespace string) (v1.PodList, bool) {
	// get PodList by Label
	podList, err := Clientset.CoreV1().Pods(namespace).List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: label,
		},
	)

	if err != nil {
		klog.Errorln("Error occured by ", label)
		klog.Errorln("Error content : ", err)
	}

	// check if podList is empty
	nilTest := []v1.Pod{}
	if reflect.DeepEqual(podList.Items, nilTest) {
		klog.Errorln(label, " cannot be found")
		return *podList, false
	}

	return *podList, true
}
