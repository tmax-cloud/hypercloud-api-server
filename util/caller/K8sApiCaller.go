package caller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	configv1alpha1 "github.com/tmax-cloud/efk-operator/api/v1alpha1"
	client "github.com/tmax-cloud/hypercloud-api-server/client"
	"github.com/tmax-cloud/hypercloud-api-server/util"
	clusterDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/cluster"
	eventDataFactory "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory/event"
	claimsv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/claim/v1alpha1"
	clusterv1alpha1 "github.com/tmax-cloud/hypercloud-multi-operator/apis/cluster/v1alpha1"
	claim "github.com/tmax-cloud/hypercloud-single-operator/api/v1alpha1"
	tmaxClusterTemplate "github.com/tmax-cloud/template-operator/api/v1"
	authApi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	eventv1 "k8s.io/api/events/v1"
	rbacApi "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/utils/pointer"
)

var Clientset *kubernetes.Clientset
var config *restclient.Config
var customClientset *client.Clientset
var AuditResourceList []string
var EventWatchChannel chan struct{}

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
	// 	klog.V(1).Infoln(err)
	// 	panic(err)
	// }
	// config.Burst = 100
	// config.QPS = 100
	// Clientset, err = kubernetes.NewForConfig(config)
	// if err != nil {
	// 	klog.V(1).Infoln(err)
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

	EventWatchChannel = make(chan struct{})
}

func GetBindableResources() map[string]string {

	type templateObjectMeta struct {
		ApiVersion string
		Kind       string
	}

	var clusterTemplates tmaxClusterTemplate.ClusterTemplateList
	var temObj templateObjectMeta

	objectList := make(map[string]string)

	data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/clustertemplates/").DoRaw(context.TODO())
	if err != nil {
		klog.V(1).Infoln(err)
		return nil
	}

	if err := json.Unmarshal(data, &clusterTemplates); err != nil {
		klog.V(1).Infoln(err)
		return nil
	}

	for _, templateItem := range clusterTemplates.Items {
		for _, objectKind := range templateItem.TemplateSpec.Objects {
			err := json.Unmarshal(objectKind.Raw, &temObj)
			if err != nil {
				klog.V(1).Infoln(err)
			} else {
				objectList[temObj.Kind] = temObj.ApiVersion
			}

		}
	}

	objectList = addBindableResources(objectList)

	return objectList
}

func addBindableResources(objectList map[string]string) map[string]string {

	includeResources := []string{"Secret", "Pod", "ReplicaSet", "DaemonSet", "Deployment", "Job", "CronJob", "StatefulSet"}
	excludeResources := []string{"Service", "Ingress", "ConfigMap",
		"Role", "RoleBinding", "ClusterRole", "ClusterRoleBinding", "Namespace", "ServiceAccount"}

	for _, resource := range excludeResources {
		_, exists := objectList[resource]
		if exists {
			delete(objectList, resource)
		}
	}

	for _, resource := range includeResources {
		if resource == "Pod" || resource == "Secret" {
			objectList[resource] = "v1"
		} else if resource == "Job" || resource == "Cronjob" {
			objectList[resource] = "batch/v1"
		} else {
			objectList[resource] = "apps/v1"
		}
	}

	objectList["Kafka"] = "kafka.strimzi.io/v1beta2"
	objectList["Redis"] = "redis.redis.opstreelabs.in/v1beta1"
	objectList["RedisCluster"] = "redis.redis.opstreelabs.in/v1beta1"

	return objectList
}

func GetNamespace(nsName string) (*corev1.Namespace, error) {
	namespace, err := Clientset.CoreV1().Namespaces().Get(context.TODO(), nsName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Info(" Namespace [ " + nsName + " ] is Not Exists")
			return nil, err
		} else {
			klog.V(3).Info("Get Namespace [ " + nsName + " ] Failed")
			klog.V(1).Infoln(err)
			return nil, err
		}
	} else {
		klog.V(3).Info("Get Namespace [ " + nsName + " ] Success")
		return namespace, nil
	}
}

func UpdateNamespace(namespace *corev1.Namespace) (*corev1.Namespace, error) {
	namespace, err := Clientset.CoreV1().Namespaces().Update(context.TODO(), namespace, metav1.UpdateOptions{})
	if err != nil {
		klog.V(3).Info("Update Namespace [ " + namespace.Name + " ] Failed")
		klog.V(1).Infoln(err)
		return nil, err
	} else {
		klog.V(3).Info("Update Namespace [ " + namespace.Name + " ] Success")
		return namespace, nil
	}
}

func CreateClusterRoleBinding(ClusterRoleBinding *rbacApi.ClusterRoleBinding) error {
	result, err := Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), ClusterRoleBinding, metav1.CreateOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}
	klog.V(3).Info(" Create ClusterRoleBinding " + result.GetObjectMeta().GetName() + " Success ")
	return nil
}

func DeleteClusterRoleBinding(name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		klog.V(1).Infoln(err)
		return err
	} else {
		klog.V(3).Info(" Delete ClusterRoleBinding " + name + " Success ")
	}
	return nil
}

func IsAccessibleNS(ns string, userId string, labelSelector string, userGroups []string) (bool, error) {
	klog.V(3).Infoln("userId : ", userId)
	for _, userGroup := range userGroups {
		klog.V(3).Infoln("userGroupName : ", userGroup)
	}

	// 1. Check If User has NS List Role
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
		klog.V(1).Infoln(err)
		return false, err
	}
	if sarResult.Status.Allowed {
		klog.V(3).Infoln(" User [ " + userId + " ] has Namespace List Role, Can Access All Namespace")
		return true, nil
	}

	// 2. Check If User has NS Get Role
	klog.V(3).Infoln(" User [ " + userId + " ] has No Namespace List Role, Check If user has Namespace Get Role to Certain Namespace")
	nsGetRuleReview := authApi.SubjectAccessReview{
		Spec: authApi.SubjectAccessReviewSpec{
			ResourceAttributes: &authApi.ResourceAttributes{
				Resource:  "namespaces",
				Verb:      "get",
				Group:     "",
				Namespace: ns,
			},
			User:   userId,
			Groups: userGroups,
		},
	}
	sarResult, err = Clientset.AuthorizationV1().SubjectAccessReviews().Create(context.TODO(), &nsGetRuleReview, metav1.CreateOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return false, err
	}
	if sarResult.Status.Allowed {
		klog.V(3).Infoln(" User [ " + userId + " ] has Namespace Get Role in Namspace [ " + ns + " ]")
		return true, nil
	}
	return false, nil
}

func GetAccessibleNS(userId string, labelSelector string, userGroups []string) (corev1.NamespaceList, error) {
	var nsList = &corev1.NamespaceList{}
	klog.V(3).Infoln("userId : ", userId)

	// // 1. Get UserGroup List if Exists
	// userDetail := getUserDetailWithoutToken(userId)
	// var userGroups []string
	// if userDetail["groups"] != nil {
	// 	for _, userGroup := range userDetail["groups"].([]interface{}) {
	// 		userGroups = append(userGroups, userGroup.(string))
	// 	}
	// }

	for _, userGroup := range userGroups {
		klog.V(3).Infoln("userGroupName : ", userGroup)
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
		klog.V(1).Infoln(err)
		return *nsList, err
	}
	if sarResult.Status.Allowed {
		klog.V(3).Infoln(" User [ " + userId + " ] has Namespace List Role, Can Access All Namespace")
		nsList, err = Clientset.CoreV1().Namespaces().List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: labelSelector,
			},
		)
		if err != nil {
			klog.V(1).Infoln(err)
			return *nsList, err
		}
	} else {
		klog.V(3).Infoln(" User [ " + userId + " ] has No Namespace List Role, Check If user has Namespace Get Role to Certain Namespace")
		potentialNsList, err := Clientset.CoreV1().Namespaces().List(
			context.TODO(),
			metav1.ListOptions{
				LabelSelector: labelSelector,
			},
		)
		if err != nil {
			klog.V(1).Infoln(err)
			return *nsList, err
		}
		var wg sync.WaitGroup
		wg.Add(len(potentialNsList.Items))
		for _, potentialNs := range potentialNsList.Items {
			go func(potentialNs corev1.Namespace, userId string, userGroups []string, nsList *corev1.NamespaceList) {
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
					klog.V(1).Infoln(err)
					panic(err)
				}
				if sarResult.Status.Allowed {
					klog.V(3).Infoln(" User [ " + userId + " ] has Namespace Get Role in Namspace [ " + potentialNs.GetName() + " ]")
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
		// 	klog.V(3).Infoln(" User [ " + userId + " ] has No Namespace Get Role in Any Namspace")
		// }
	}
	// if len(nsList.Items) > 0 {
	// 	klog.V(3).Infoln("=== [ " + userId + " ] Accessible Namespace ===")
	// 	for _, ns := range nsList.Items {
	// 		klog.V(3).Infoln("  " + ns.Name)
	// 	}
	// }
	return *nsList, nil
}

// var nsList = &corev1.NamespaceList{}
func GetAccessibleNSC(userId string, userGroups []string, labelSelector string) (claim.NamespaceClaimList, error) {
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
		klog.V(1).Infoln(err)
		return *nscList, err
	}
	// klog.V(3).Infoln("sarResult : " + sarResult.String())

	// /apis/claim.tmax.io/v1alpha1/namespaceclaims?labelselector
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").Param(util.QUERY_PARAMETER_LABEL_SELECTOR, labelSelector).DoRaw(context.TODO())
	if err != nil {
		klog.V(1).Infoln(err)
		return *nscList, err
	}

	if sarResult.Status.Allowed {
		klog.V(3).Infoln(" User [ " + userId + " ] has NamespaceClaim List Role, Can Access All NamespaceClaim")

		if err := json.Unmarshal(data, &nscList); err != nil {
			klog.V(1).Infoln(err)
			return *nscList, err
		}

	} else {
		klog.V(3).Infoln(" User [ " + userId + " ] has No NamespaceClaim List Role, Check If user has NamespaceClaim Get Role & has Owner Annotation on certain NamespaceClaim")
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
			klog.V(1).Infoln(err)
			return *nscList, err
		}
		if sarResult.Status.Allowed {
			klog.V(3).Infoln(" User [ " + userId + " ] has NamespaceClaim Get Role")
			var potentialNscList = &claim.NamespaceClaimList{}
			if err := json.Unmarshal(data, &potentialNscList); err != nil {
				klog.V(1).Infoln(err)
				return *nscList, err
			}

			var wg sync.WaitGroup
			wg.Add(len(potentialNscList.Items))
			for _, potentialNsc := range potentialNscList.Items {
				go func(potentialNsc claim.NamespaceClaim, userId string, nscList *claim.NamespaceClaimList) {
					defer wg.Done()
					if potentialNsc.Annotations["owner"] == userId {
						klog.V(3).Infoln(" User [ " + userId + " ] has owner annotation in NamspaceClaim [ " + potentialNsc.Name + " ]")
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
			// 	klog.V(3).Infoln(" User [ " + userId + " ] has No owner annotation in Any NamspaceClaim")
			// }
		} else {
			klog.V(3).Infoln(" User [ " + userId + " ] has no NamespaceClaim Get Role, User Cannot Access any NamespaceClaim")
		}

	}

	if len(nscList.Items) > 0 {
		klog.V(3).Infoln("=== [ " + userId + " ] Accessible NamespaceClaim ===")
		// for _, nsc := range nscList.Items {
		// 	klog.V(3).Infoln("  ", nsc.Name)
		// }
	}
	return *nscList, nil
}

func DeleteRQCWithUser(userId string) error {
	var rqcList = &claim.ResourceQuotaClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/").Namespace("").Resource("resourcequotaclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/").Namespace("").Resource("resourcequotaclaims").DoRaw(context.TODO()) // for hypercloud5 version
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	if err = json.Unmarshal(data, &rqcList); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	for _, rqc := range rqcList.Items {
		if rqc.Annotations["creator"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(rqc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.V(1).Infoln(err)
				klog.V(3).Infoln("Faile to delete ResourceQuotaClaim ", rqc.Name)
				continue
			}
			klog.V(3).Infoln("ResourceQuotaClaim ", rqc.Name, " is deleted")
		}
	}
	return nil
}

func DeleteNSCWithUser(userId string) error {
	var nscList = &claim.NamespaceClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/namespaceclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/namespaceclaims").DoRaw(context.TODO()) // for hypercloud5 version
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	if err = json.Unmarshal(data, &nscList); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	for _, nsc := range nscList.Items {
		if nsc.Annotations["owner"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(nsc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.V(1).Infoln(err)
				klog.V(3).Infoln("Faile to delete NamespaceClaim ", nsc.Name)
				continue
			}
			klog.V(3).Infoln("NamespaceClaim ", nsc.Name, " is deleted")
		}
	}
	return nil
}

func DeleteRBCWithUser(userId string) error {
	var rbcList = &claim.RoleBindingClaimList{}
	//data, err := Clientset.RESTClient().Get().AbsPath("/apis/tmax.io/v1/").Namespace("").Resource("rolebindingclaims").DoRaw(context.TODO()) // for hypercloud4 version
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/claim.tmax.io/v1alpha1/").Namespace("").Resource("rolebindingclaims").DoRaw(context.TODO()) // for hypercloud5 version
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	if err = json.Unmarshal(data, &rbcList); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	for _, rbc := range rbcList.Items {
		if rbc.Annotations["creator"] == userId {
			_, err := Clientset.RESTClient().Delete().AbsPath(rbc.SelfLink).DoRaw(context.TODO())
			if err != nil {
				klog.V(1).Infoln(err)
				klog.V(3).Infoln("Faile to delete RoleBindingClaim ", rbc.Name)
				continue
			}
			klog.V(3).Infoln("RoleBindingClaim ", rbc.Name, " is deleted")
		}
	}
	return nil
}

func DeleteCRBWithUser(userId string) error {
	crbList, err := Clientset.RbacV1().ClusterRoleBindings().List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	for _, crb := range crbList.Items {
		for _, subject := range crb.Subjects {
			if subject.Name == userId {
				err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.V(1).Infoln(err)
					klog.V(3).Infoln("Faile to delete ClusterRoleBinding ", crb.ObjectMeta.Name)
				} else {
					klog.V(3).Infoln("ClusterRoleBinding ", crb.ObjectMeta.Name, " is deleted")
				}
			}
		}
	}
	return nil
}

func GetCRBAdmin() string {
	crbList, err := Clientset.RbacV1().ClusterRoleBindings().List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		klog.V(1).Infoln(err)
	}
	var adminemail string
	for _, crb := range crbList.Items {
		if crb.Name == "admin" {
			adminemail = crb.Subjects[0].Name
			klog.V(3).Infof("admin is " + adminemail)
		}
	}
	return adminemail
}

func DeleteRBWithUser(userId string) error {
	rbList, err := Clientset.RbacV1().RoleBindings("").List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	for _, rb := range rbList.Items {
		for _, subject := range rb.Subjects {
			if subject.Name == userId {
				err := Clientset.RbacV1().RoleBindings(rb.ObjectMeta.Namespace).Delete(context.TODO(), rb.ObjectMeta.Name, metav1.DeleteOptions{})
				if err != nil {
					klog.V(1).Infoln(err)
					klog.V(3).Infoln("Faile to delete RoleBinding ", rb.ObjectMeta.Name)
				} else {
					klog.V(3).Infoln("RoleBinding", rb.ObjectMeta.Name, "is deleted")
				}
			}
		}
	}
	return nil
}

// ExecCommand sends a 'exec' command to specific pod.
// It returns outputs of command.
// If the container parameter == "", it chooses first container.
func ExecCommand(pod corev1.Pod, command []string, container string) (string, string, error) {

	var stdin io.Reader

	req := Clientset.CoreV1().RESTClient().Post().Resource("pods").Name(pod.Name).
		Namespace(pod.Namespace).SubResource("exec")

	if container == "" {
		container = pod.Spec.Containers[0].Name
	}

	option := &corev1.PodExecOptions{
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
func GetPodListByLabel(label string, namespace string) (corev1.PodList, bool) {
	// get PodList by Label
	podList, err := Clientset.CoreV1().Pods(namespace).List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: label,
		},
	)

	if err != nil {
		klog.V(1).Infoln("Error occurred by " + label)
		klog.V(1).Infoln("Error content : " + err.Error())
		return *podList, false
	}

	// check if podList is empty
	nilTest := []corev1.Pod{}
	if reflect.DeepEqual(podList.Items, nilTest) {
		klog.V(1).Infoln(label, " cannot be found")
		return *podList, false
	}

	return *podList, true
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
		klog.V(1).Infoln(err)
		return nil, err
	}

	return sarResult, nil
}

func AdmitClusterClaim(userId string, userGroups []string, clusterClaim *claimsv1alpha1.ClusterClaim, admit bool, reason string) (*claimsv1alpha1.ClusterClaim, error) {

	clusterClaimStatusUpdateRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims/status", clusterClaim.Namespace, clusterClaim.Name, "update")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clusterClaimStatusUpdateRuleResult.Status.Allowed {
		klog.V(3).Infoln(" User [ " + userId + " ] has ClusterClaims/status Update Role, Can Update ClusterClaims")

		if admit {
			clusterClaim.Status.Phase = "Approved"
			if reason == "" {
				clusterClaim.Status.Reason = "Administrator approve the claim"
			} else {
				clusterClaim.Status.Reason = reason
			}
		} else {
			clusterClaim.Status.Phase = "Rejected"
			if reason == "" {
				clusterClaim.Status.Reason = "Administrator reject the claim"
			} else {
				clusterClaim.Status.Reason = reason
			}
		}

		result, err := customClientset.ClaimsV1alpha1().ClusterClaims(clusterClaim.Namespace).
			UpdateStatus(context.TODO(), clusterClaim, metav1.UpdateOptions{})
		if err != nil {
			klog.V(1).Infoln("Update ClusterClaim [ " + clusterClaim.Name + " ] Failed")
			return nil, err
		} else {
			msg := "Update ClusterClaim [ " + clusterClaim.Name + " ] Success"
			klog.V(3).Infoln(msg)
			return result, nil
		}
	} else {
		newErr := errors.NewBadRequest("User [ " + userId + " ] has No ClusterClaims/status Update Role, Check If user has ClusterClaims/status Update Role")
		klog.V(1).Infoln(newErr)
		return nil, newErr
	}
}

func GetClusterClaim(userId string, userGroups []string, clusterClaimName string, clusterClaimNamespace string) (*claimsv1alpha1.ClusterClaim, error) {

	var clusterClaim = &claimsv1alpha1.ClusterClaim{}

	clusterClaimGetRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims", clusterClaimNamespace, clusterClaimName, "get")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clusterClaimGetRuleResult.Status.Allowed {
		clusterClaim, err = customClientset.ClaimsV1alpha1().ClusterClaims(clusterClaimNamespace).Get(context.TODO(), clusterClaimName, metav1.GetOptions{})
		if err != nil {
			klog.V(1).Infoln(err)
			return nil, err
		}
	} else {
		newErr := errors.NewBadRequest("User [" + userId + "] authorization is denied for clusterclaims [" + clusterClaimName + "]")
		klog.V(1).Infoln(newErr)
		return nil, newErr
	}

	return clusterClaim, nil
}

func ListAllClusterClaims(userId string, userGroups []string) (*claimsv1alpha1.ClusterClaimList, error) {
	var clusterClaimList = &claimsv1alpha1.ClusterClaimList{}

	clusterClaimListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims", "", "", "list")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	clusterClaimList, err = customClientset.ClaimsV1alpha1().ClusterClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}
	clusterClaimList.Kind = "ClusterClaimList"
	clusterClaimList.APIVersion = "claims.tmax.io/v1alpha1"

	if clusterClaimListRuleResult.Status.Allowed {
		msg := "User [ " + userId + " ] has ClusterClaim List Role, Can Access All ClusterClaim"
		klog.V(3).Infoln(msg)
		if len(clusterClaimList.Items) == 0 {
			msg := "No ClusterClaim was Found."
			klog.V(3).Infoln(msg)
			return clusterClaimList, nil
		}
		return clusterClaimList, nil
	} else {
		msg := "User [ " + userId + " ] has No permission to list clusterclaims on all namespaces"
		klog.V(3).Infoln(msg)
		clusterClaimList.Items = []claimsv1alpha1.ClusterClaim{}
		return clusterClaimList, nil
	}
}

// cluster created type check, cluster owner만 승인/거절 할 수 있도록 check, ready 상태의 cluster만 허용하도록 check
func CheckClusterValid(userId string, clusterName string, cucNamespace string) error {
	clusterManager, err := customClientset.ClusterV1alpha1().ClusterManagers(cucNamespace).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		klog.V(1).Info(err)
		return err
	}

	if clusterManager.GetClusterType() != clusterv1alpha1.ClusterTypeCreated {
		errMsg := fmt.Sprintf("cluster[ %s ]'s cluster type is not created", clusterManager.Name)
		klog.V(1).Info(errMsg)
		return fmt.Errorf(errMsg)
	}

	clusterOwner, ok := clusterManager.Annotations["creator"]
	if !ok {
		errMsg := fmt.Sprintf("cannot check cluster[ %s ]'s owner. missing cluster manager annotation[creator].", clusterManager.Name)
		klog.V(1).Info(errMsg)
		return fmt.Errorf(errMsg)
	}

	if clusterManager.Status.Phase != clusterv1alpha1.ClusterManagerPhaseReady {
		errMsg := fmt.Sprintf("cluster[ %s ] is not ready", clusterManager.Name)
		klog.V(1).Info(errMsg)
		return fmt.Errorf(errMsg)
	}

	if clusterOwner != userId {
		errMsg := fmt.Sprintf("userId[ %s ] is not cluster owner.", userId)
		klog.V(1).Info(errMsg)
		return fmt.Errorf(errMsg)
	}
	return nil
}

func ListAccessibleClusterClaims(userId string, userGroups []string, namespace string) (*claimsv1alpha1.ClusterClaimList, error) {
	var clusterClaimList = &claimsv1alpha1.ClusterClaimList{}

	clusterClaimList.Kind = "ClusterClaimList"
	clusterClaimList.APIVersion = "claims.tmax.io/v1alpha1"

	clusterClaimListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterclaims", namespace, "", "list")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}
	if clusterClaimListRuleResult.Status.Allowed {
		clusterClaimList, err = customClientset.ClaimsV1alpha1().ClusterClaims(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.V(1).Info(err)
		}
		klog.V(3).Infoln("Success list clusterclaim in namespace [ " + namespace + " ]")
		if len(clusterClaimList.Items) == 0 {
			klog.V(3).Infoln(" User [ " + userId + " ] has No ClusterClaim")
		}
		return clusterClaimList, nil
	} else {
		klog.V(3).Infoln("User [ " + userId + " ] has No permission in namespace  [ " + namespace + " ]")
		clusterClaimList.Items = []claimsv1alpha1.ClusterClaim{}
		return clusterClaimList, nil
	}

}

func ListClusterUpdateClaims(userId string, userGroups []string) (*claimsv1alpha1.ClusterUpdateClaimList, error) {
	return ListClusterUpdateClaimsByNamespace(userId, userGroups, "")
}

func ListClusterUpdateClaimsByNamespace(userId string, userGroups []string, namespace string) (*claimsv1alpha1.ClusterUpdateClaimList, error) {

	clusterUpdateClaimListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterupdateclaims", namespace, "", "list")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clusterUpdateClaimListRuleResult.Status.Allowed {
		clusterUpdateClaimList, err := customClientset.ClaimsV1alpha1().ClusterUpdateClaims(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			klog.V(1).Info(err)
		}
		if namespace == "" {
			klog.V(3).Infoln("Success list clusterupdateclaim")
		} else {
			klog.V(3).Infoln(fmt.Sprintf("Success list clusterupdateclaim in namespace [ %s ]", namespace))
		}
		if len(clusterUpdateClaimList.Items) == 0 {
			klog.V(3).Infoln(fmt.Sprintf("User [ %s ] has No ClusterUpdateClaim", userId))
		}
		return clusterUpdateClaimList, nil
	} else {
		if namespace == "" {
			klog.V(3).Infoln(fmt.Sprintf("User [ %s ] has No permission", userId))
		} else {
			klog.V(3).Infoln(fmt.Sprintf("User [ %s ] has No permission in namespace  [ %s ]", userId, namespace))
		}
		return &claimsv1alpha1.ClusterUpdateClaimList{}, nil
	}
}

func GetClusterUpdateClaim(userId string, userGroups []string, cucName string, cucNamespace string) (*claimsv1alpha1.ClusterUpdateClaim, error) {
	clusterUpdateClaimGetRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterupdateclaims", cucNamespace, cucName, "get")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	var cuc *claimsv1alpha1.ClusterUpdateClaim
	if clusterUpdateClaimGetRuleResult.Status.Allowed {
		cuc, err = customClientset.ClaimsV1alpha1().ClusterUpdateClaims(cucNamespace).Get(context.TODO(), cucName, metav1.GetOptions{})
		if err != nil {
			klog.V(1).Infoln(err)
			return nil, err
		}
	} else {
		newErr := errors.NewBadRequest("User [" + userId + "] authorization is denied for clusterupdateclaims [" + cucName + "]")
		klog.V(1).Infoln(newErr)
		return nil, newErr
	}
	return cuc, nil
}

func AdmitClusterUpdateClaim(userId string, userGroups []string, cuc *claimsv1alpha1.ClusterUpdateClaim, admit bool, reason string) (*claimsv1alpha1.ClusterUpdateClaim, error) {
	clusterUpdateClaimStatusRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLAIM_API_GROUP, "clusterupdateclaims/status", cuc.Namespace, cuc.Name, "update")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clusterUpdateClaimStatusRuleResult.Status.Allowed {
		msg := fmt.Sprintf(" User [ %s ] has ClusterUpdateClaims/status Update Role, Can Update ClusterUpdateClaims", userId)
		klog.V(3).Infoln(msg)
		if admit {
			cuc.Status.Phase = claimsv1alpha1.ClusterUpdateClaimPhaseApproved
		} else {
			cuc.Status.Phase = claimsv1alpha1.ClusterUpdateClaimPhaseRejected
			if reason == "" {
				cuc.Status.Reason = claimsv1alpha1.ClusterUpdateClaimReason("Administrator rejected the claim")
			} else {
				cuc.Status.Reason = claimsv1alpha1.ClusterUpdateClaimReason(reason)
			}
		}

		result, err := customClientset.
			ClaimsV1alpha1().
			ClusterUpdateClaims(cuc.Namespace).
			UpdateStatus(context.TODO(), cuc, metav1.UpdateOptions{})
		if err != nil {
			msg := fmt.Sprintf("Update ClusterUpdateClaim [ %s ] failed", cuc.Name)
			klog.V(1).Infoln(msg)
			return nil, err
		} else {
			msg := fmt.Sprintf("Update ClusterUpdateClaim [ %s ] success", cuc.Name)
			klog.V(3).Infoln(msg)
			return result, nil
		}
	} else {
		msg := fmt.Sprintf("User [ %s ] has No ClusterUpdateClaims/status Update Role, Check If user has ClusterUpdateClaims/status Update Role", userId)
		newErr := errors.NewBadRequest(msg)
		klog.V(1).Infoln(newErr)
		return nil, newErr
	}
}

func ListAllCluster(userId string, userGroups []string) (*clusterv1alpha1.ClusterManagerList, error) {
	var clmList = &clusterv1alpha1.ClusterManagerList{}

	clmListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", "", "list")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	clmList, err = customClientset.ClusterV1alpha1().ClusterManagers("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}
	clmList.Kind = "ClusterManagerList"
	clmList.APIVersion = "cluster.tmax.io/v1alpha1"

	if clmListRuleResult.Status.Allowed {
		msg := "User [ " + userId + " ] has ClusterManager List Role, Can Access All ClusterManager"
		klog.V(3).Infoln(msg)
		if len(clmList.Items) == 0 {
			msg := "No cluster was Found."
			klog.V(3).Infoln(msg)
		}
		return clmList, nil
	} else {
		return ListAccessibleCluster(userId, userGroups)
		// msg := "User [ " + userId + " ] has No permission to list ClusterManager on all namespaces"
		// klog.V(3).Infoln(msg)
		// clmList.Items = []clusterv1alpha1.ClusterManager{}
		// return clmList, nil
	}
}

func ListAccessibleCluster(userId string, userGroups []string) (*clusterv1alpha1.ClusterManagerList, error) {

	var clmList = &clusterv1alpha1.ClusterManagerList{}

	clmList.Kind = "ClusterManagerList"
	clmList.APIVersion = "cluster.tmax.io/v1alpha1"

	clmList, err := customClientset.ClusterV1alpha1().ClusterManagers("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	NamespacedNameList, err := clusterDataFactory.ListClusterAllNamespace(userId, userGroups)
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	_clmList := util.Search(NamespacedNameList, clmList)

	if len(_clmList.Items) == 0 {
		msg := " User [ " + userId + " ] has No Clusters"
		klog.V(3).Infoln(msg)
		return _clmList, nil
	}
	msg := " User [ " + userId + " ] has Clusters"
	klog.V(3).Infoln(msg)
	return _clmList, nil
}

func ListClusterInNamespace(userId string, userGroups []string, namespace string) (*clusterv1alpha1.ClusterManagerList, error) {

	var clmList = &clusterv1alpha1.ClusterManagerList{}

	clmList.Kind = "ClusterManagerList"
	clmList.APIVersion = "cluster.tmax.io/v1alpha1"

	clmListRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", namespace, "", "list")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	clmList, err = customClientset.ClusterV1alpha1().ClusterManagers(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clmListRuleResult.Status.Allowed {
		if len(clmList.Items) == 0 {
			msg := " User [ " + userId + " ] has No clusterManager"
			klog.V(3).Infoln(msg)
			return clmList, nil
		}
		msg := "Success list cluster in a namespace [ " + namespace + " ]"
		klog.V(3).Infoln(msg)
		return clmList, nil
	} else {
		// ns에 list 권한 없으면 db에서 속한것만 찾아서 반환!
		// db에서 읽어온다.
		clusterNameList, err := clusterDataFactory.ListClusterInNamespace(userId, userGroups, namespace)
		if err != nil {
			klog.V(1).Infoln(err)
			return nil, err
		}
		_clmList := []clusterv1alpha1.ClusterManager{}

		for _, clm := range clmList.Items {
			if util.Contains(clusterNameList, clm.Name) {
				_clmList = append(_clmList, clm)
			}
		}
		clmList.Items = _clmList
		if len(clmList.Items) == 0 {
			msg := " User [ " + userId + " ] has No Clusters"
			klog.V(3).Infoln(msg)
			return clmList, nil
		}
		msg := " User [ " + userId + " ] has Clusters"
		klog.V(3).Infoln(msg)
		return clmList, nil
	}

}

func GetCluster(userId string, userGroups []string, clusterName string, namespace string) (*clusterv1alpha1.ClusterManager, error) {
	var clm = &clusterv1alpha1.ClusterManager{}
	clusterGetRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", namespace, clusterName, "get")
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}

	if clusterGetRuleResult.Status.Allowed {
		clm, err = customClientset.ClusterV1alpha1().ClusterManagers(namespace).Get(context.TODO(), clusterName, metav1.GetOptions{})
		if err != nil {
			klog.V(1).Infoln(err)
			return nil, err
		}
	} else {
		newErr := errors.NewBadRequest("User [" + userId + "] authorization is denied for cluster [" + clusterName + "]")
		klog.V(1).Infoln(newErr.Error())
		return nil, newErr
	}

	return clm, nil
}

func GetClusterWithoutSAR(userId string, userGroups []string, clusterName string, namespace string) (*clusterv1alpha1.ClusterManager, error) {
	clm, err := customClientset.ClusterV1alpha1().ClusterManagers(namespace).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		klog.V(1).Infoln(err)
		return nil, err
	}
	return clm, nil
}

func CheckClusterManagerDuplication(clusterName string, namespace string) (bool, error) {
	if _, err := customClientset.ClusterV1alpha1().ClusterManagers(namespace).Get(context.TODO(), clusterName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		} else {
			return true, err
		}
	} else {
		return true, nil
	}

}

func CreateCLMRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	var roleName string
	var roleBindingName string
	roleBinding := &rbacApi.RoleBinding{}
	if attribute == "user" {
		roleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		roleBindingName = subject + "-user-" + clusterManager.Name + "-clm-rolebinding"
		roleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		roleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		roleBindingName = subject + "-group-" + clusterManager.Name + "-clm-rolebinding"
		roleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	role := &rbacApi.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: clusterManager.Namespace,
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
		Rules: []rbacApi.PolicyRule{
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clustermanagers"},
				ResourceNames: []string{clusterManager.Name}, Verbs: []string{"get"}},
			{APIGroups: []string{util.CLUSTER_API_GROUP}, Resources: []string{"clustermanagers/status"},
				ResourceNames: []string{clusterManager.Name}, Verbs: []string{"get"}},
		},
	}

	if _, err := Clientset.RbacV1().Roles(clusterManager.Namespace).Create(context.TODO(), role, metav1.CreateOptions{}); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	roleBinding.ObjectMeta = metav1.ObjectMeta{
		Name:      roleBindingName,
		Namespace: clusterManager.Namespace,
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
	}
	roleBinding.RoleRef = rbacApi.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "Role",
		Name:     roleName,
	}

	if _, err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{}); err != nil {
		klog.V(1).Infoln(err)
		return err
	}
	msg := "ClusterManager role [" + roleName + "] and rolebinding [ " + roleBindingName + "]  is created"
	klog.V(3).Infoln(msg)

	return nil
}

// master cluster에 생성한 cluster manager role, rolebinding 삭제
func DeleteCLMRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {

	var roleName string
	var roleBindingName string
	if attribute == "user" {
		roleName = subject + "-user-" + clusterManager.Name + "-clm-role"
		roleBindingName = subject + "-user-" + clusterManager.Name + "-clm-rolebinding"
	} else {
		roleName = subject + "-group-" + clusterManager.Name + "-clm-role"
		roleBindingName = subject + "-group-" + clusterManager.Name + "-clm-rolebinding"
	}

	if _, err := Clientset.RbacV1().Roles(clusterManager.Namespace).Get(context.TODO(), roleName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infoln("Role [" + roleName + "] is already deleted. pass")
			return nil
		} else {
			klog.V(1).Infoln("Error: Get clusterrole [" + roleName + "] is failed")
			return err
		}
	} else {
		if err := Clientset.RbacV1().Roles(clusterManager.Namespace).Delete(context.TODO(), roleName, metav1.DeleteOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	}

	if _, err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Get(context.TODO(), roleBindingName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infoln("Rolebinding [" + roleBindingName + "] is already deleted. pass")
			return nil
		} else {
			klog.V(1).Infoln("Error: Get clusterrole [" + roleName + "] is failed")
			return err
		}
	} else {
		if err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Delete(context.TODO(), roleBindingName, metav1.DeleteOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	}

	return nil

}

// defunct
// func GetConsoleService(namespace string, name string) (*corev1.Service, error) {
// 	result, err := Clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
// 	return result, err
// }

func GetFbc(namespace string, name string) (*configv1alpha1.FluentBitConfiguration, error) {
	result, err := customClientset.ConfigV1alpha1().FluentBitConfigurations(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return result, err
}

func CreateNSGetRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	clusterRoleName := "clusterrole-ns-get"
	var roleBindingName string
	roleBinding := &rbacApi.RoleBinding{}

	if attribute == "user" {
		// clusterRoleName = subject + "-user-" + clusterManager.Namespace + "-" + clusterManager.Name + "-clusterrole"
		roleBindingName = subject + "-user-ns-get-rolebinding"
		roleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "User",
				Name:     subject,
			},
		}
	} else {
		// clusterRoleName = subject + "-group-" + clusterManager.Namespace + "-" + clusterManager.Name + "-clusterrole"
		roleBindingName = subject + "-group-ns-get-rolebinding"
		roleBinding.Subjects = []rbacApi.Subject{
			{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Group",
				Name:     subject,
			},
		}
	}

	roleBinding.ObjectMeta = metav1.ObjectMeta{
		Name:      roleBindingName,
		Namespace: clusterManager.Namespace,
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
	}
	roleBinding.RoleRef = rbacApi.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     clusterRoleName,
	}

	if _, err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			msg := "User [" + subject + "] already has a namespace get rolebinding in a namespace [" + clusterManager.Namespace + "]"
			klog.V(3).Info(msg)
			return nil
		} else {
			klog.V(1).Infoln(err)
			return err
		}
	}

	msg := "Namespace Get Rolebinding [ " + roleBindingName + "]  is created"
	klog.V(3).Infoln(msg)

	return nil
}

// master cluster에 생성한 namespace rolebinding 삭제
func DeleteNSGetRole(clusterManager *clusterv1alpha1.ClusterManager, subject string, attribute string) error {
	var roleBindingName string
	if attribute == "user" {
		roleBindingName = subject + "-user-ns-get-rolebinding"
	} else {
		roleBindingName = subject + "-group-ns-get-rolebinding"
	}

	// Subject가 해당 ns에 사용중인 클러스터가 남았다면 ns get rolebinding 삭제 안하고.. 없으면 삭제한다. (이전에 db에서 현재 요청에대한 클러스터는 삭제함)
	if res, err := clusterDataFactory.GetRemainClusterForSubject(clusterManager.Namespace, subject, attribute); err != nil {
		klog.V(1).Infoln(err)
		return err
	} else if res != 0 {
		klog.V(3).Info("User [" + subject + "] has a remain cluster in a namespace [" + clusterManager.Namespace + "].. do not delete ns-get-rolebinding")
		return nil
	} else {
		if _, err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Get(context.TODO(), roleBindingName, metav1.GetOptions{}); err != nil {
			if errors.IsNotFound(err) {
				klog.V(3).Infoln("Rolebinding [" + roleBindingName + "] is already deleted. pass")
				return nil
			} else {
				klog.V(1).Infoln("Error: Get clusterrole [" + roleBindingName + "] is failed")
				return err
			}
		} else {
			if err := Clientset.RbacV1().RoleBindings(clusterManager.Namespace).Delete(context.TODO(), roleBindingName, metav1.DeleteOptions{}); err != nil {
				klog.V(1).Infoln(err)
				return err
			}
		}
	}

	return nil
}

func WatchK8sEvent() {

	watchlist := cache.NewListWatchFromClient(Clientset.EventsV1().RESTClient(), "events", "", fields.Everything())

	_, controller := cache.NewInformer(
		watchlist,
		&eventv1.Event{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				e := obj.(*eventv1.Event)
				eventDataFactory.Insert(e)
			},
			DeleteFunc: func(obj interface{}) {
				e := obj.(*eventv1.Event)
				eventDataFactory.Insert(e)
			},
			UpdateFunc: func(olde, newe interface{}) {
				e := newe.(*eventv1.Event)
				eventDataFactory.Insert(e)
			},
		},
	)

	EventWatchChannel = make(chan struct{})
	go controller.Run(EventWatchChannel)
}

func UpdateAuditResourceList() {
	AuditResourceList = []string{"users"}
	tmp := make(map[string]struct{})
	fullName := make(map[string]struct{})
	apiGroupList := &metav1.APIGroupList{}
	data, err := Clientset.RESTClient().Get().AbsPath("/apis/").DoRaw(context.TODO())
	if err != nil {
		klog.V(1).Infoln(err)
	}
	if err := json.Unmarshal(data, apiGroupList); err != nil {
		klog.V(1).Infoln(err)
	}

	for _, apiGroup := range apiGroupList.Groups {
		for _, version := range apiGroup.Versions {
			apiResourceList := &metav1.APIResourceList{}
			path := strings.Replace("/apis/{GROUPVERSION}", "{GROUPVERSION}", version.GroupVersion, -1)
			data, err := Clientset.RESTClient().Get().AbsPath(path).DoRaw(context.TODO())
			if err != nil {
				klog.V(1).Infoln(err)
			}
			if err := json.Unmarshal(data, apiResourceList); err != nil {
				klog.V(1).Infoln(err)
			}

			for _, apiResource := range apiResourceList.APIResources {
				fullName[apiResource.Name] = struct{}{}
				if !strings.Contains(apiResource.Name, "/") {
					if _, ok := tmp[apiResource.Name]; !ok {
						tmp[apiResource.Name] = struct{}{}
					}
				}
			}
		}
	}

	//corev1 group
	apiResourceList := &metav1.APIResourceList{}
	data, err = Clientset.RESTClient().Get().AbsPath("/api/v1").DoRaw(context.TODO())
	if err != nil {
		klog.V(1).Infoln(err)
	}
	if err := json.Unmarshal(data, apiResourceList); err != nil {
		klog.V(1).Infoln(err)
	}
	for _, apiResource := range apiResourceList.APIResources {
		fullName[apiResource.Name] = struct{}{}
		if !strings.Contains(apiResource.Name, "/") {
			if _, ok := tmp[apiResource.Name]; !ok {
				tmp[apiResource.Name] = struct{}{}
			}
		}
	}

	// map to string
	for k := range tmp {
		AuditResourceList = append(AuditResourceList, k)
	}
}

func GetHyperAuthAdminAccount() (string, string, error) {
	secret := &corev1.Secret{}
	var err error

	if secret, err = Clientset.CoreV1().Secrets("hyperauth").Get(context.TODO(), "passwords", metav1.GetOptions{}); errors.IsNotFound(err) {
		klog.V(1).Infoln("Hyperauth password secret is not found")
		return "", "", err
	} else if err != nil {
		klog.V(1).Infoln(err, "Failed to get hyperauth password secret")
		return "", "", err
	}

	id := string(secret.Data["HYPERAUTH_ADMIN"])
	password := string(secret.Data["HYPERAUTH_PASSWORD"])

	return id, password, nil
}

// kubectlInit create 'hypercloud-kubectl' namespace
// and role/rolebinding for command 'exec' to the pod
func kubectlInit(userName string) error {

	if _, err := Clientset.CoreV1().Namespaces().Get(context.TODO(), util.HYPERCLOUD_KUBECTL_NAMESPACE, metav1.GetOptions{}); errors.IsNotFound(err) {
		var ns corev1.Namespace
		ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: util.HYPERCLOUD_KUBECTL_NAMESPACE,
				Labels: map[string]string{
					"hypercloud": "system", // for avoding hypercloud-mutator webhook configuration
				},
			},
		}
		if _, err := Clientset.CoreV1().Namespaces().Create(context.TODO(), &ns, metav1.CreateOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	} else if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	RoleNameForPod := ParseUserName(userName) + "-exec"

	if _, err := Clientset.RbacV1().Roles(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), RoleNameForPod+"-role", metav1.GetOptions{}); errors.IsNotFound(err) {
		var role rbacApi.Role
		role = rbacApi.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      RoleNameForPod + "-role",
				Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
				Labels: map[string]string{
					util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
				},
			},
			Rules: []rbacApi.PolicyRule{
				{
					APIGroups: []string{
						"", // corev1
					},
					Resources: []string{
						"pods/exec",
						"pods",
					},
					ResourceNames: []string{
						util.HYPERCLOUD_KUBECTL_PREFIX + ParseUserName(userName),
					},
					Verbs: []string{
						"get",
					},
				},
			},
		}
		if _, err := Clientset.RbacV1().Roles(util.HYPERCLOUD_KUBECTL_NAMESPACE).Create(context.TODO(), &role, metav1.CreateOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	} else if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	if _, err := Clientset.RbacV1().RoleBindings(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), RoleNameForPod+"-rolebinding", metav1.GetOptions{}); errors.IsNotFound(err) {
		var rolebinding rbacApi.RoleBinding
		rolebinding = rbacApi.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      RoleNameForPod + "-rolebinding",
				Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
				Labels: map[string]string{
					util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
				},
			},
			RoleRef: rbacApi.RoleRef{
				Kind:     "Role",
				APIGroup: "rbac.authorization.k8s.io",
				Name:     RoleNameForPod + "-role",
			},
			Subjects: []rbacApi.Subject{
				{
					Kind: "User",
					Name: userName,
					// Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
				},
			},
		}
		if _, err := Clientset.RbacV1().RoleBindings(util.HYPERCLOUD_KUBECTL_NAMESPACE).Create(context.TODO(), &rolebinding, metav1.CreateOptions{}); err != nil {
			klog.V(1).Infoln(err)
			return err
		}
	} else if err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	return nil
}

// DeployKubectlPod makes serviceaccount which has same authorization compared to given userName(email),
// then deploy pod with kubectl image
func DeployKubectlPod(userName string) error {
	if err := kubectlInit(userName); err != nil {
		klog.V(1).Infoln(err)
		return err
	}
	kubectlName := util.HYPERCLOUD_KUBECTL_PREFIX + ParseUserName(userName)

	// If pod already exists,
	// Delete it if the status of pod is completed(Succeeded)
	// Do nothing if the status of pod is not completed(Running)
	if pod, err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), kubectlName, metav1.GetOptions{}); err == nil {
		if pod.Status.Phase != "Running" {
			if err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), kubectlName, metav1.DeleteOptions{}); err != nil {
				return err
			}
		} else {
			return nil // errors.NewBadRequest("Pod [" + kubectlName + "] already exists")
		}
	}

	// Create ServiceAccount if not exists
	if _, err := Clientset.CoreV1().ServiceAccounts(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), kubectlName, metav1.GetOptions{}); errors.IsNotFound(err) {
		var sa corev1.ServiceAccount
		sa = corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
				Name:      kubectlName,
				Labels: map[string]string{
					util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
				},
			},
		}
		if _, err := Clientset.CoreV1().ServiceAccounts(util.HYPERCLOUD_KUBECTL_NAMESPACE).Create(context.TODO(), &sa, metav1.CreateOptions{}); err != nil {
			return err
		}
	}

	// Make New Rolebinding and ClusterRoleBinding for ServiceAccount
	type rolebindingSubject struct {
		Kind string
		Name []string
	}
	group, err := GetHyperAuthGroupByUser(userName)
	if err != nil {
		klog.V(1).Infoln(err)
		return err
	}
	subjectGroup := rolebindingSubject{"Group", group}
	subjectUser := rolebindingSubject{"User", []string{userName}}

	if err := CreateRBForKubectlSA(kubectlName, subjectGroup, subjectUser); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	if err := CreateCRBForKubectlSA(kubectlName, subjectGroup, subjectUser); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	// Make configmap to change default namespace in kubectl container
	var configmapName string
	if configmapName, err = CreateConfigmapForKubectl(kubectlName, 1); err != nil {
		klog.V(1).Infoln(err)
		return err
	}

	// Deploy kubectl Pod using generated ServiceAccount
	sleepTime := os.Getenv("KUBECTL_TIMEOUT")
	if len(sleepTime) == 0 || sleepTime == "{KUBECTL_TIMEOUT}" {
		sleepTime = "21600" // 6 hours
	}
	var kubectlPod corev1.Pod
	kubectlPod = corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubectlName,
			Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
			Labels: map[string]string{
				util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: kubectlName,
			DNSPolicy:          "ClusterFirst",
			RestartPolicy:      "Never",
			Containers: []corev1.Container{
				{
					Image: util.HYPERCLOUD_KUBECTL_IMAGE,
					Name:  "kubectl",
					Command: []string{
						"/bin/sh", "-c",
					},
					Args: []string{
						"sleep " + sleepTime,
						//"tail -f /dev/null;",
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "default-namespace",
							MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "default-namespace",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configmapName,
							},
						},
					},
				},
			},
		},
	}
	if _, err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Create(context.TODO(), &kubectlPod, metav1.CreateOptions{}); err != nil {
		return err
	}

	// Delete configmap after pod runs for security
	count := 0
	for count < util.HYPERCLOUD_KUBECTL_CONFIGMAP_DELETE_WAIT_TIME {
		if pod, err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), kubectlName, metav1.GetOptions{}); err == nil {
			if pod.Status.Phase == "Running" {
				configmapName := kubectlName + "-configmap"
				if err := Clientset.CoreV1().ConfigMaps(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), configmapName, metav1.DeleteOptions{}); err != nil {
					klog.V(1).Infoln(err)
					count++
					time.Sleep(time.Second * 1)
				} else {
					klog.V(3).Infoln("Delete Configmap [" + configmapName + "] Success")
					break
				}
			} else {
				count++
				time.Sleep(time.Second * 1)
			}
		} else {
			count++
			time.Sleep(time.Second * 1)
		}
	}

	if count == util.HYPERCLOUD_KUBECTL_CONFIGMAP_DELETE_WAIT_TIME {
		klog.V(1).Infoln("Failed to delete Configmap [" + configmapName + "]. ca.crt is exposed by configmap, please delete it manually")
	}

	return nil
}

func ParseUserName(userName string) string {
	str := strings.Replace(userName, "_", "-", -1)
	return strings.Replace(str, "@", ".", -1)
}

func CreateRBForKubectlSA(saName string, S ...struct {
	Kind string
	Name []string
}) error {

	rbList, err := Clientset.RbacV1().RoleBindings("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var newRB rbacApi.RoleBinding
	for _, rb := range rbList.Items {
		for _, sub := range rb.Subjects {
			for _, s := range S {
				if sub.Kind == s.Kind {
					for _, n := range s.Name {
						if sub.Name == n {
							// if already exists, delete original one and create
							if _, err := Clientset.RbacV1().RoleBindings(rb.ObjectMeta.Namespace).Get(context.TODO(), util.HYPERCLOUD_KUBECTL_PREFIX+rb.ObjectMeta.Name, metav1.GetOptions{}); err == nil {
								if err := Clientset.RbacV1().RoleBindings(rb.ObjectMeta.Namespace).Delete(context.TODO(), util.HYPERCLOUD_KUBECTL_PREFIX+rb.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
									klog.V(1).Infoln(err)
								}
							}
							newRB = rbacApi.RoleBinding{
								TypeMeta: rb.TypeMeta,
								ObjectMeta: metav1.ObjectMeta{
									Name: saName + "-" + rb.ObjectMeta.Name,
									//Namespace:       rb.ObjectMeta.Namespace,
									Labels: map[string]string{
										util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
									},
									Annotations:     rb.ObjectMeta.Annotations,
									OwnerReferences: rb.ObjectMeta.OwnerReferences,
								},
								Subjects: []rbacApi.Subject{
									{
										Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
										Kind:      "ServiceAccount",
										Name:      saName,
									},
								},
								RoleRef: rb.RoleRef,
							}
							if _, err := Clientset.RbacV1().RoleBindings(rb.ObjectMeta.Namespace).Create(context.TODO(), &newRB, metav1.CreateOptions{}); err != nil {
								klog.V(1).Infoln(err)
							} else {
								klog.V(3).Infoln("Create RoleBinding [" + saName + "-" + rb.ObjectMeta.Name + "] Success")
							}
						}
					}

				}
			}
		}
	}
	return nil
}
func CreateCRBForKubectlSA(saName string, S ...struct {
	Kind string
	Name []string
}) error {
	crbList, err := Clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	var newCRB rbacApi.ClusterRoleBinding
	for _, crb := range crbList.Items {
		for _, sub := range crb.Subjects {
			for _, s := range S {
				if sub.Kind == s.Kind {
					for _, n := range s.Name {
						if sub.Name == n {
							// if already exists, delete original one and create
							if _, err := Clientset.RbacV1().ClusterRoleBindings().Get(context.TODO(), util.HYPERCLOUD_KUBECTL_PREFIX+crb.ObjectMeta.Name, metav1.GetOptions{}); err == nil {
								if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), util.HYPERCLOUD_KUBECTL_PREFIX+crb.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
									klog.V(1).Infoln(err)
								}
							}
							newCRB = rbacApi.ClusterRoleBinding{
								TypeMeta: crb.TypeMeta,
								ObjectMeta: metav1.ObjectMeta{
									Name:      saName + "-" + crb.ObjectMeta.Name,
									Namespace: crb.ObjectMeta.Namespace,
									Labels: map[string]string{
										util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
									},
									Annotations:     crb.ObjectMeta.Annotations,
									OwnerReferences: crb.ObjectMeta.OwnerReferences,
								},
								Subjects: []rbacApi.Subject{
									{
										Namespace: util.HYPERCLOUD_KUBECTL_NAMESPACE,
										Kind:      "ServiceAccount",
										Name:      saName,
									},
								},
								RoleRef: crb.RoleRef,
							}
							if _, err := Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), &newCRB, metav1.CreateOptions{}); err != nil {
								klog.V(1).Infoln(err)
							} else {
								klog.V(3).Infoln("Create ClusterRoleBinding [" + saName + "-" + crb.ObjectMeta.Name + "] Success")
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// CreateConfigmapForKubectl creates configmap for volume mounting at /var/run/secrets/kubernetes.io/serviceaccount
// to change default namespace of kubectl container.
func CreateConfigmapForKubectl(serviceAccountName string, retry int) (string, error) {
	if retry >= 10 {
		return "", errors.NewServiceUnavailable("Serviceaccount [" + serviceAccountName + "] is not created")
	}

	var sa *corev1.ServiceAccount
	var err error
	if sa, err = Clientset.CoreV1().ServiceAccounts(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), serviceAccountName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			klog.V(3).Infoln("Wait for Serviceaccount [" + serviceAccountName + "] to be created...")
			time.Sleep(time.Second * 1)
			return CreateConfigmapForKubectl(serviceAccountName, retry+1)
		} else {
			return "", errors.NewServiceUnavailable("Failed to get Serviceaccount [" + serviceAccountName + "]")
		}
	}

	secretName := sa.Secrets[0].Name
	configmapName := serviceAccountName + "-configmap"
	if secret, err := Clientset.CoreV1().Secrets(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), secretName, metav1.GetOptions{}); err != nil {
		return "", errors.NewServiceUnavailable("Failed to get Secret [" + secretName + "]")
	} else {
		caCrt := string(secret.Data["ca.crt"])
		token := string(secret.Data["token"])
		namespace := "default"

		if _, err := Clientset.CoreV1().ConfigMaps(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), configmapName, metav1.GetOptions{}); err == nil {
			if err := Clientset.CoreV1().ConfigMaps(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), configmapName, metav1.DeleteOptions{}); err != nil {
				klog.V(1).Infoln(err)
			}
		}

		if _, err := Clientset.CoreV1().ConfigMaps(util.HYPERCLOUD_KUBECTL_NAMESPACE).Create(context.TODO(), &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configmapName,
				Labels: map[string]string{
					util.HYPERCLOUD_KUBECTL_LABEL_KEY: util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
				},
			},
			Data: map[string]string{
				"ca.crt":    caCrt,
				"token":     token,
				"namespace": namespace,
			},
		}, metav1.CreateOptions{}); err != nil {
			klog.V(3).Infoln("Failed to create Configmap for Serviceaccount [" + serviceAccountName + "]")
			return "", err
		} else {
			klog.V(3).Infoln("Create Configmap [" + configmapName + "]")
		}
	}

	return configmapName, nil
}

// DeleteKubectlResource deletes all kubectl pod related resources for give userName,
// which contains Pod, RoleBinding, ClusterRoleBinding and ServiceAccount
func DeleteKubectlResourceByUserName(userName string) error {
	kubectlName := util.HYPERCLOUD_KUBECTL_PREFIX + ParseUserName(userName)

	if _, err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), kubectlName, metav1.GetOptions{}); err == nil {
		if err := Clientset.CoreV1().Pods(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), kubectlName, metav1.DeleteOptions{}); err != nil {
			klog.V(1).Infoln(err)
		}
		klog.V(3).Infoln("Delete Pod [" + kubectlName + "] Success")
	}

	if _, err := Clientset.CoreV1().ServiceAccounts(util.HYPERCLOUD_KUBECTL_NAMESPACE).Get(context.TODO(), kubectlName, metav1.GetOptions{}); err == nil {
		if err := Clientset.CoreV1().ServiceAccounts(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), kubectlName, metav1.DeleteOptions{}); err != nil {
			klog.V(1).Infoln(err)
		}
		klog.V(3).Infoln("Delete ServiceAccount [" + kubectlName + "] Success")
	}

	rbList, err := Clientset.RbacV1().RoleBindings("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: util.HYPERCLOUD_KUBECTL_LABEL_KEY + "=" + util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
	})
	if err != nil {
		return err
	}
	for _, rb := range rbList.Items {
		for _, sub := range rb.Subjects {
			if (sub.Name == kubectlName && sub.Kind == "ServiceAccount") || rb.Name == ParseUserName(userName)+"-exec"+"-rolebinding" {
				if err := Clientset.RbacV1().RoleBindings(rb.Namespace).Delete(context.TODO(), rb.Name, metav1.DeleteOptions{}); err != nil {
					klog.V(1).Infoln(err)
				}
				klog.V(3).Infoln("Delete RoleBinding [" + rb.Name + "] Success")
				break
			}
		}
	}
	crbList, err := Clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), metav1.ListOptions{
		LabelSelector: util.HYPERCLOUD_KUBECTL_LABEL_KEY + "=" + util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
	})
	if err != nil {
		return err
	}
	for _, crb := range crbList.Items {
		for _, sub := range crb.Subjects {
			if sub.Name == kubectlName && sub.Kind == "ServiceAccount" {
				if err := Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), crb.Name, metav1.DeleteOptions{}); err != nil {
					klog.V(1).Infoln(err)
				}
				klog.V(3).Infoln("Delete ClusterRoleBinding [" + crb.Name + "] Success")
				break
			}
		}
	}

	if err := Clientset.RbacV1().Roles(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), ParseUserName(userName)+"-exec"+"-role", metav1.DeleteOptions{}); err != nil {
		klog.V(1).Infoln(err)
	} else {
		klog.V(3).Infoln("Delete Role [" + ParseUserName(userName) + "-exec" + "-role" + "] Success")
	}

	if err := Clientset.CoreV1().ConfigMaps(util.HYPERCLOUD_KUBECTL_NAMESPACE).Delete(context.TODO(), kubectlName+"-configmap", metav1.DeleteOptions{}); err != nil {
		klog.V(1).Infoln(err)
	} else {
		klog.V(3).Infoln("Delete ConfigMap [" + kubectlName + "] Success")
	}

	return nil
}

// DeleteKubectlResource deletes all kubectl pod related resources for all user,
// which contains Pod, RoleBinding, ClusterRoleBinding and ServiceAccount.
// It only runs by cronJob, not by calling API.
func DeleteKubectlAllResource() {
	klog.V(4).Infoln("Start Garbage Collect Kubectl Related Resources")
	var DeletedUserName []string
	if podList, err := Clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: util.HYPERCLOUD_KUBECTL_LABEL_KEY + "=" + util.HYPERCLOUD_KUBECTL_LABEL_VALUE,
	}); err != nil {
		klog.V(1).Infoln(err)
		return
	} else {
		for _, pod := range podList.Items {
			if pod.Status.Phase != "Running" {
				split := strings.Index(pod.Name, "-kubectl-")
				userName := pod.Name[split+9:]
				DeletedUserName = append(DeletedUserName, userName)
				if err := Clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
					klog.V(1).Infoln(err)
				}
				klog.V(3).Infoln("Delete Pod [" + pod.Name + "] Success")
			}
		}
	}

	if len(DeletedUserName) < 1 {
		klog.V(4).Infoln("No garbage resources for kubectl")
		return
	}

	for _, userName := range DeletedUserName {
		if err := DeleteKubectlResourceByUserName(userName); err != nil {
			klog.V(1).Infoln(err)
		}
	}

	klog.V(4).Infoln("Complete Garbage Collect Kubectl Related Resources")
}
