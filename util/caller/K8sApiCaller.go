package caller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"io"
	"reflect"
	"sync"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/pointer"

	claimsv1alpha1 "github.com/tmax-cloud/claim-operator/api/v1alpha1"
	clusterv1alpha1 "github.com/tmax-cloud/cluster-manager-operator/api/v1alpha1"
	configv1alpha1 "github.com/tmax-cloud/efk-operator/api/v1alpha1"
	alertModel "github.com/tmax-cloud/hypercloud-api-server/alert/model"
	client "github.com/tmax-cloud/hypercloud-api-server/client"
	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	authApi "k8s.io/api/authorization/v1"
	coreApi "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacApi "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

var Clientset *kubernetes.Clientset
var config *restclient.Config
var customClientset *client.Clientset

func init() {

	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig2", filepath.Join(home, ".kube", "config"), "/root/.kube")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig2", "", "/root/.kube")
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
	config, err = restclient.InClusterConfig()
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
	customClientset, err = client.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

}

func GetNamespace(nsName string) *v1.Namespace {
	namespace, err := Clientset.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(" Namespace [ " + nsName + " ] is Not Exists")
			return nil
		} else {
			klog.Info("Get Namespace [ " + nsName + " ] Failed")
			klog.Errorln(err)
			panic(err)
		}
	} else {
		klog.Info("Get Namespace [ " + nsName + " ] Success")
		return namespace
	}
}

func UpdateNamespace(namespace *v1.Namespace) *v1.Namespace {
	namespace, err := Clientset.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
	if err != nil {
		klog.Info("Update Namespace [ " + namespace.Name + " ] Failed")
		klog.Errorln(err)
		panic(err)
	} else {
		klog.Info("Update Namespace [ " + namespace.Name + " ] Success")
		return namespace
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

func GetAccessibleNS(userId string, labelSelector string, userGroups []string) coreApi.NamespaceList {
	var nsList = &coreApi.NamespaceList{}
	klog.Infoln("userId : ", userId)

	// // 1. Get UserGroup List if Exists
	// userDetail := getUserDetailWithoutToken(userId)
	// var userGroups []string
	// if userDetail["groups"] != nil {
	// 	for _, userGroup := range userDetail["groups"].([]interface{}) {
	// 		userGroups = append(userGroups, userGroup.(string))
	// 	}
	// }

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
		klog.Infoln(" User [ " + userId + " ] has Namespace List Role, Can Access All Namespace")
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
		klog.Infoln(" User [ " + userId + " ] has No Namespace List Role, Check If user has Namespace Get Role to Certain Namespace")
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
					klog.Infoln(" User [ " + userId + " ] has Namespace Get Role in Namspace [ " + potentialNs.GetName() + " ]")
					nsList.Items = append(nsList.Items, potentialNs)
				}
			}(potentialNs, userId, userGroups, nsList)
		}
		wg.Wait()

		// if len(nsList.Items) > 0 {
		nsList.APIVersion = potentialNsList.APIVersion
		nsList.Continue = potentialNsList.Continue
		nsList.ResourceVersion = potentialNsList.ResourceVersion
		nsList.TypeMeta = potentialNsList.TypeMeta
		// } else {
		// 	klog.Infoln(" User [ " + userId + " ] has No Namespace Get Role in Any Namspace")
		// }
	}
	// if len(nsList.Items) > 0 {
	// 	klog.Infoln("=== [ " + userId + " ] Accessible Namespace ===")
	// 	for _, ns := range nsList.Items {
	// 		klog.Infoln("  " + ns.Name)
	// 	}
	// }
	return *nsList
}

// var nsList = &coreApi.NamespaceList{}
func GetAccessibleNSC(userId string, userGroups []string, labelSelector string) claim.NamespaceClaimList {
	var nscList = &claim.NamespaceClaimList{}

	// 1. Check If User has NSC List Role
	nsListRuleReview := authApi.SubjectAccessReview{
		Spec: authApi.SubjectAccessReviewSpec{
			ResourceAttributes: &authApi.ResourceAttributes{
				Resource: "namespaceclaims",
				Verb:     "list",
				Group:    util.HYPERCLOUD4_CLAIM_API_GROUP,
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
	// klog.Infoln("sarResult : " + sarResult.String())

	// /apis/claim.tmax.io/v1alpha1/namespaceclaims?labelselector
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").Param(util.QUERY_PARAMETER_LABEL_SELECTOR, labelSelector).DoRaw(context.TODO())
	if err != nil {
		klog.Errorln(err)
		panic(err)
	}

	if sarResult.Status.Allowed {
		klog.Infoln(" User [ " + userId + " ] has NamespaceClaim List Role, Can Access All NamespaceClaim")

		if err := json.Unmarshal(data, &nscList); err != nil {
			klog.Errorln(err)
			panic(err)
		}

	} else {
		klog.Infoln(" User [ " + userId + " ] has No NamespaceClaim List Role, Check If user has NamespaceClaim Get Role & has Owner Annotation on certain NamespaceClaim")
		// 2. Check If User has NSC Get Role
		nscGetRuleReview := authApi.SubjectAccessReview{
			Spec: authApi.SubjectAccessReviewSpec{
				ResourceAttributes: &authApi.ResourceAttributes{
					Resource: "namespaceclaims",
					Verb:     "get",
					Group:    util.HYPERCLOUD4_CLAIM_API_GROUP,
				},
				User:   userId,
				Groups: userGroups,
			},
		}

		sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nscGetRuleReview, metav1.CreateOptions{})
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
		if sarResult.Status.Allowed {
			klog.Infoln(" User [ " + userId + " ] has NamespaceClaim Get Role")
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
						klog.Infoln(" User [ " + userId + " ] has owner annotation in NamspaceClaim [ " + potentialNsc.Name + " ]")
						nscList.Items = append(nscList.Items, potentialNsc)
					}
				}(potentialNsc, userId, nscList)
			}
			wg.Wait()

			// if len(nscList.Items) > 0 {
			nscList.APIVersion = potentialNscList.APIVersion
			nscList.Continue = potentialNscList.Continue
			nscList.ResourceVersion = potentialNscList.ResourceVersion
			nscList.TypeMeta = potentialNscList.TypeMeta
			// } else {
			// 	klog.Infoln(" User [ " + userId + " ] has No owner annotaion in Any NamspaceClaim")
			// }
		} else {
			klog.Infoln(" User [ " + userId + " ] has no NamespaceClaim Get Role, User Cannot Access any NamespaceClaim")
		}

	}

	if len(nscList.Items) > 0 {
		klog.Infoln("=== [ " + userId + " ] Accessible NamespaceClaim ===")
		// for _, nsc := range nscList.Items {
		// 	klog.Infoln("  ", nsc.Name)
		// }
	}
	return *nscList
}

func DeleteRQCWithUser(userId string) {
	var rqcList = &claim.ResourceQuotaClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/").Namespace("").Resource("resourcequotaclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/").Namespace("").Resource("resourcequotaclaims").DoRaw(context.TODO()) // for hypercloud5 version

	if err = json.Unmarshal(data, &rqcList); err != nil {
		klog.Errorln(err)
		panic(err)
	}

	for _, rqc := range rqcList.Items {
		if rqc.Annotations["creator"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(rqc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.Errorln(err)
				panic(err)
			}
		}
	}
	klog.Infoln("Successfully delete ResourceQuotaClaim made by", userId)
}

func DeleteNSCWithUser(userId string) {
	var nscList = &claim.NamespaceClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/namespaceclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").DoRaw(context.TODO()) // for hypercloud5 version

	if err = json.Unmarshal(data, &nscList); err != nil {
		klog.Errorln(err)
		panic(err)
	}

	for _, nsc := range nscList.Items {
		if nsc.Annotations["creator"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(nsc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.Errorln(err)
				panic(err)
			}
		}
	}
	klog.Infoln("Successfully delete NamespaceClaim made by", userId)
}

func DeleteRBCWithUser(userId string) {
	var rbcList = &claim.RoleBindingClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/").Namespace("").Resource("rolebindingclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/").Namespace("").Resource("rolebindingclaims").DoRaw(context.TODO()) // for hypercloud5 version

	if err = json.Unmarshal(data, &rbcList); err != nil {
		klog.Errorln(err)
		panic(err)
	}

	for _, rbc := range rbcList.Items {
		if rbc.Annotations["creator"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(rbc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.Errorln(err)
				panic(err)
			}
		}
	}
	klog.Infoln("Successfully delete RoleBindingClaim made by", userId)
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
		klog.Errorln("Error occured by " + label)
		klog.Errorln("Error content : " + err.Error())
		return *podList, false
	}

	// check if podList is empty
	nilTest := []v1.Pod{}
	if reflect.DeepEqual(podList.Items, nilTest) {
		klog.Errorln(label, " cannot be found")
		return *podList, false
	}

	return *podList, true
}
func CreateAlert(body alertModel.Alert, ns string) {
	bodyByte, _ := json.Marshal(body)
	// data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").Param(util.QUERY_PARAMETER_LABEL_SELECTOR, labelSelector).DoRaw(context.TODO())

	data := Clientset.RESTClient().Post().AbsPath("/apis/tmax.io/v1/namespaces/" + ns + "/alerts").Body(bodyByte).Do(context.TODO())
	klog.Infoln(data)
	// if err != nil {
	//    klog.Errorln(err)
	//    panic(err)
	// }
	//return data
}

func GetAlert(name string, ns string, label string) alertModel.Alert {
	var u alertModel.Alert
	// data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").Param(util.QUERY_PARAMETER_LABEL_SELECTOR, labelSelector).DoRaw(context.TODO())
	if name != "" {
		data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/namespaces/"+ns+"/alerts").Param("name", name).DoRaw(context.TODO())
		klog.Infoln(data)
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}

		json.Unmarshal([]byte(data), &u)

	} else if label != "" {
		data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/namespaces/"+ns+"/alerts").Param("label", label).DoRaw(context.TODO())
		klog.Infoln(data)
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
		json.Unmarshal([]byte(data), &u)

	} else {
		data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/namespaces/" + ns + "/alerts").DoRaw(context.TODO())
		klog.Infoln(data)
		if err != nil {
			klog.Errorln(err)
			panic(err)
		}
		json.Unmarshal([]byte(data), &u)

	}

	return u
}

func CreateSubjectAccessReview(userId string, userGroups []string, group string, resource string, namespace string, name string, verb string) (*authApi.SubjectAccessReview, error) {
	sar := &authApi.SubjectAccessReview{
		Spec: authApi.SubjectAccessReviewSpec{
			ResourceAttributes: &authApi.ResourceAttributes{
				Group:     group,
				Resource:  resource,
				Namespace: namespace,
				Name:      name,
				Verb:      verb,
			},
			User:   userId,
			Groups: userGroups,
		},
	}

	sarResult, err := Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return sarResult, nil
}

func AdmitClusterClaim(userId string, userGroups []string, clusterClaim *claimsv1alpha1.ClusterClaim, admit bool, reason string) (*claimsv1alpha1.ClusterClaim, string, int) {

	clusterClaimStatusUpdateRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims/status", "", "", "update")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	if clusterClaimStatusUpdateRuleResult.Status.Allowed {
		klog.Infoln(" User [ " + userId + " ] has ClusterClaims/status Update Role, Can Update ClusterClaims")

		if admit == true {
			clusterClaim.Status.Phase = "Admitted"
			if reason == "" {
				clusterClaim.Status.Reason = "Administrator approve the claim"
			} else {
				clusterClaim.Status.Reason = reason
			}
		} else {
			clusterClaim.Status.Phase = "Rejected"
			if reason == "" {
				clusterClaim.Status.Reason = "Administrator approve the claim"
			} else {
				clusterClaim.Status.Reason = reason
			}
		}

		result, err := customClientset.ClaimsV1alpha1().ClusterClaims().
			UpdateStatus(context.TODO(), clusterClaim, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorln("Update ClusterClaim [ " + clusterClaim.Name + " ] Failed")
			return nil, err.Error(), http.StatusInternalServerError
		} else {
			msg := "Update ClusterClaim [ " + clusterClaim.Name + " ] Success"
			klog.Infoln(msg)
			return result, msg, http.StatusOK
		}
	} else {
		msg := " User [ " + userId + " ] has No ClusterClaims/status Update Role, Check If user has ClusterClaims/status Update Role"
		klog.Infoln(msg)
		return nil, msg, http.StatusForbidden
	}
}

func GetClusterClaim(userId string, userGroups []string, clusterClaimName string) (*claimsv1alpha1.ClusterClaim, string, int) {

	var clusterClaim = &claimsv1alpha1.ClusterClaim{}

	clusterClaimGetRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims", "", clusterClaimName, "get")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	if clusterClaimGetRuleResult.Status.Allowed {
		clusterClaim, err = customClientset.ClaimsV1alpha1().ClusterClaims().Get(context.TODO(), clusterClaimName, metav1.GetOptions{})
		if err != nil {
			klog.Errorln(err)
			return nil, err.Error(), http.StatusInternalServerError
		}
	} else {
		msg := "User [" + userId + "] authorization is denied for clusterclaims [" + clusterClaimName + "]"
		klog.Infoln(msg)
		return nil, msg, http.StatusForbidden
	}

	return clusterClaim, "Get claim success", http.StatusOK
}

func ListAccessibleClusterClaims(userId string, userGroups []string) (*claimsv1alpha1.ClusterClaimList, string, int) {
	var clusterClaimList = &claimsv1alpha1.ClusterClaimList{}

	clusterClaimListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims", "", "", "list")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	clusterClaimList, err = customClientset.ClaimsV1alpha1().ClusterClaims().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}
	clusterClaimList.Kind = "ClusterClaimList"
	clusterClaimList.APIVersion = "claims.tmax.io/v1alpha1"

	if clusterClaimListRuleResult.Status.Allowed {
		msg := "User [ " + userId + " ] has ClusterClaim List Role, Can Access All ClusterClaim"
		klog.Infoln(msg)
		if len(clusterClaimList.Items) == 0 {
			msg := "No ClusterClaim was Found."
			klog.Infoln(msg)
		}
		return clusterClaimList, msg, http.StatusOK
	} else {
		klog.Infoln(" User [ " + userId + " ] has No ClusterClaim List Role, Check If user has ClusterClaim Get Role & has Owner Annotation on certain ClusterClaim")
		_clusterClaimList := []claimsv1alpha1.ClusterClaim{}
		// var wg sync.WaitGroup
		// wg.Add(len(clusterClaimList.Items))
		for _, clusterClaim := range clusterClaimList.Items {
			// go func(clusterClaim claimsv1alpha1.ClusterClaim, userId string, _clusterClaimList []claimsv1alpha1.ClusterClaim) {
			// defer wg.Done()
			if clusterClaim.Annotations["creator"] == userId {
				klog.Infoln(" User [ " + userId + " ] has owner annotation in ClusterClaim [ " + clusterClaim.Name + " ]")
				_clusterClaimList = append(_clusterClaimList, clusterClaim)
			}
			// }(clusterClaim, userId, _clusterClaimList)
		}
		// wg.Wait()

		clusterClaimList.Items = _clusterClaimList

		if len(clusterClaimList.Items) == 0 {
			msg := " User [ " + userId + " ] has No ClusterClaim"
			klog.Infoln(msg)
			return clusterClaimList, msg, http.StatusOK
		}
	}
	msg := "Success to get ClusterClaim for User [ " + userId + " ]"
	klog.Infoln(msg)
	return clusterClaimList, msg, http.StatusOK
}

func ListCluster(userId string, userGroups []string, accessible bool) (*clusterv1alpha1.ClusterManagerList, string, int) {

	var clmList = &clusterv1alpha1.ClusterManagerList{}

	clmListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clusterclaims", "", "", "list")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	clmList, err = customClientset.ClusterV1alpha1().ClusterManagers().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}
	clmList.Kind = "ClusterManagerList"
	clmList.APIVersion = "cluster.tmax.io/v1alpha1"

	if clmListRuleResult.Status.Allowed && !accessible {
		msg := "User [ " + userId + " ] has ClusterManager List Role, Can Access All ClusterManager"
		klog.Infoln(msg)
		if len(clmList.Items) == 0 {
			msg := "No cluster was Found."
			klog.Infoln(msg)
		}
		return clmList, msg, http.StatusOK
	} else {
		klog.Infoln("User [ " + userId + " ] has No ClusterManager List Role or acessible is true")
		_clmList := []clusterv1alpha1.ClusterManager{}
		for _, clm := range clmList.Items {
			if _, ok := clm.Status.Owner[userId]; ok {
				// if clm.Status.Owner == userId {
				_clmList = append(_clmList, clm)
			} else if _, ok := clm.Status.Members[userId]; ok {
				// } else if util.Contains(clm.Status.Members, userId) {
				_clmList = append(_clmList, clm)
			} else {
				for _, userGroup := range userGroups {
					if _, ok := clm.Status.Groups[userGroup]; ok {
						// if util.Contains(clm.Status.Groups, userGroup) {
						_clmList = append(_clmList, clm)
						break
					}
				}
			}
		}
		clmList.Items = _clmList

		if len(clmList.Items) == 0 {
			msg := " User [ " + userId + " ] has No Clusters"
			klog.Infoln(msg)
			return clmList, msg, http.StatusOK
		}
		msg := " User [ " + userId + " ] has Clusters"
		klog.Infoln(msg)
		return clmList, msg, http.StatusOK
	}
}

func GetCluster(userId string, userGroups []string, clusterName string) (*clusterv1alpha1.ClusterManager, string, int) {

	var clm = &clusterv1alpha1.ClusterManager{}
	clusterGetRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", clusterName, "get")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	if clusterGetRuleResult.Status.Allowed {
		clm, err = customClientset.ClusterV1alpha1().ClusterManagers().Get(context.TODO(), clusterName, metav1.GetOptions{})
		if err != nil {
			klog.Errorln(err)
			return nil, err.Error(), http.StatusInternalServerError
		}
	} else {
		msg := "User [" + userId + "] authorization is denied for cluster [" + clusterName + "]"
		klog.Infoln(msg)
		return nil, msg, http.StatusForbidden
	}
	return clm, "Get cluster success", http.StatusOK
}

func UpdateClusterManager(userId string, userGroups []string, clm *clusterv1alpha1.ClusterManager) (*clusterv1alpha1.ClusterManager, string, int) {

	clmUpdateRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", clm.Name, "update")
	if err != nil {
		klog.Errorln(err)
		return nil, err.Error(), http.StatusInternalServerError
	}

	if clmUpdateRuleResult.Status.Allowed {
		result, err := customClientset.ClusterV1alpha1().ClusterManagers().UpdateStatus(context.TODO(), clm, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorln("Update member list in cluster [ " + clm.Name + " ] Failed")
			return nil, err.Error(), http.StatusInternalServerError
		} else {
			msg := "Update member list in cluster [ " + clm.Name + " ] Success"
			klog.Infoln(msg)
			return result, msg, http.StatusOK
		}
	} else {
		msg := " User [ " + userId + " ] is not a cluster admin, Cannot invite members"
		klog.Infoln(msg)
		return nil, msg, http.StatusForbidden
	}
}

func CreateCLMRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, isGroup bool) (string, int) {
	var clusterRoleName string
	var clusterRoleBindingName string
	clusterRoleBinding := &rbacApi.ClusterRoleBinding{}
	if !isGroup {
		clusterRoleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-user-" + clusterManager.Name + "-clm-rolebinding"
		clusterRoleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		clusterRoleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-group-" + clusterManager.Name + "-clm-rolebinding"
		clusterRoleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	clusterRole := &rbacApi.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion:         util.CLUSTER_API_GROUP_VERSION,
					Kind:               util.CLUSTER_API_Kind,
					Name:               clusterManager.GetName(),
					UID:                clusterManager.GetUID(),
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				},
			},
		},
		Rules: []rbacApi.PolicyRule{
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clustermanagers"},
				ResourceNames: []string{clusterManager.Name}, Verbs: []string{"get"}},
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clustermanagers/status"},
				ResourceNames: []string{clusterManager.Name}, Verbs: []string{"get"}},
		},
	}

	if _, err := Clientset.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}

	clusterRoleBinding.ObjectMeta = metav1.ObjectMeta{
		Name: clusterRoleBindingName,
		OwnerReferences: []metav1.OwnerReference{
			metav1.OwnerReference{
				APIVersion:         util.CLUSTER_API_GROUP_VERSION,
				Kind:               util.CLUSTER_API_Kind,
				Name:               clusterManager.GetName(),
				UID:                clusterManager.GetUID(),
				BlockOwnerDeletion: pointer.BoolPtr(true),
				Controller:         pointer.BoolPtr(true),
			},
		},
	}
	clusterRoleBinding.RoleRef = rbacApi.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
	}

	if _, err := Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}
	msg := "ClusterMnager role [" + clusterRoleName + "] and rolebinding [ " + clusterRoleBindingName + "]  is created"
	klog.Infoln(msg)

	return msg, http.StatusOK
}

func DeleteCLMRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, isGroup bool) (string, int) {

	var clusterRoleName string
	var clusterRoleBindingName string
	if !isGroup {
		clusterRoleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-user-" + clusterManager.Name + "-clm-rolebinding"
	} else {
		clusterRoleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		clusterRoleBindingName = subject + "-group-" + clusterManager.Name + "-clm-rolebinding"
	}

	_, err := Clientset.RbacV1().ClusterRoles().Get(context.TODO(), clusterRoleName, metav1.GetOptions{})
	if err == nil {
		if err := Clientset.RbacV1().ClusterRoles().Delete(context.TODO(), clusterRoleName, metav1.DeleteOptions{}); err != nil {
			klog.Errorln(err)
			return err.Error(), http.StatusInternalServerError
		}
	} else if errors.IsNotFound(err) {
		klog.Infoln("Role [" + clusterRoleName + "] is already deleted. pass")
		return err.Error(), http.StatusOK
	} else {
		klog.Errorln("Error: Get clusterrole [" + clusterRoleName + "] is failed")
		return err.Error(), http.StatusInternalServerError
	}

	_, err = Clientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
	if err == nil {
		if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{}); err != nil {
			klog.Errorln(err)
			return err.Error(), http.StatusInternalServerError
		}
	} else if errors.IsNotFound(err) {
		klog.Infoln("Rolebinding [" + clusterRoleBindingName + "] is already deleted. pass")
		return err.Error(), http.StatusOK
	} else {
		klog.Errorln("Error: Get clusterrole [" + clusterRoleName + "] is failed")
		return err.Error(), http.StatusInternalServerError
	}
	msg := "ClusterMnager role [" + clusterRoleName + "] and rolebinding [ " + clusterRoleBindingName + "]  is deleted"
	klog.Infoln(msg)

	return msg, http.StatusOK
}

func GetFbc(namespace string, name string) (*configv1alpha1.FluentBitConfiguration, error) {
	result, err := customClientset.ConfigV1alpha1().FluentBitConfigurations(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return result, err
}

func CreateClusterClaim(userId string, userGroup []string, cc *claimsv1alpha1.ClusterClaim) (*claimsv1alpha1.ClusterClaim, string, int) {
	if len(cc.Annotations) == 0 {
		cc.Annotations = map[string]string{
			"owner": userId,
		}
	} else {
		cc.Annotations["owner"] = userId
	}
	result, err := customClientset.ClaimsV1alpha1().ClusterClaims().Create(context.TODO(), cc, metav1.CreateOptions{})
	if err != nil {
		klog.Errorln("Create ClusterClaim  [ " + cc.Name + " ] Failed")
		return nil, err.Error(), http.StatusInternalServerError
	}

	msg := "Create ClusterClaim  [ " + cc.Name + " ] Success"
	klog.Infoln(msg)
	return result, msg, http.StatusOK

}
func CreateCCRole(userId string, userGroup []string, cc *claimsv1alpha1.ClusterClaim) (string, int) {
	claimRoleName := userId + "-" + cc.Name + "-cc-role"
	clusterRole := &rbacApi.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: claimRoleName,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion:         util.CLAIM_API_GROUP_VERSION,
					Kind:               util.CLAIM_API_Kind,
					Name:               cc.GetName(),
					UID:                cc.GetUID(),
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				},
			},
		},
		Rules: []rbacApi.PolicyRule{
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clusterclaims"},
				ResourceNames: []string{cc.Name}, Verbs: []string{"get"}},
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clusterclaims/status"},
				ResourceNames: []string{cc.Name}, Verbs: []string{"get"}},
		},
	}

	if _, err := Clientset.RbacV1().ClusterRoles().Create(context.TODO(), clusterRole, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}

	claimRoleBindingName := userId + "-" + cc.Name + "-cc-rolebinding"
	clusterRoleBinding := &rbacApi.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: claimRoleBindingName,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion:         util.CLAIM_API_GROUP_VERSION,
					Kind:               util.CLAIM_API_Kind,
					Name:               cc.GetName(),
					UID:                cc.GetUID(),
					BlockOwnerDeletion: pointer.BoolPtr(true),
					Controller:         pointer.BoolPtr(true),
				},
			},
		},
		RoleRef: rbacApi.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     claimRoleName,
		},
		Subjects: []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     userId,
			},
		},
	}

	if _, err := Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), clusterRoleBinding, metav1.CreateOptions{}); err != nil {
		klog.Errorln(err)
		return err.Error(), http.StatusInternalServerError
	}
	msg := "ClusterClaim role [" + claimRoleName + "] and rolebinding [ " + claimRoleBindingName + "]  is created"
	klog.Infoln(msg)

	return msg, http.StatusOK
}
