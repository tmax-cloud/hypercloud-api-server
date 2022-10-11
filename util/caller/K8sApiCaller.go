package caller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

// func UpdateAuditResourceList() {
// 	AuditResourceList = make(map[string][]string)
// 	apiGroupList := &metav1.APIGroupList{}
// 	data, err := Clientset.RESTClient().Get().AbsPath("/apis/").DoRaw(context.TODO())
// 	if err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}
// 	if err := json.Unmarshal(data, apiGroupList); err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}

// 	for _, apiGroup := range apiGroupList.Groups {
// 		ListAPIResource(&apiGroup)
// 	}

// 	apiResourceList := &metav1.APIResourceList{}
// 	data, err = Clientset.RESTClient().Get().AbsPath("/api/v1").DoRaw(context.TODO())
// 	if err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}
// 	if err := json.Unmarshal(data, apiResourceList); err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}
// 	for _, apiResource := range apiResourceList.APIResources {
// 		if !strings.Contains(apiResource.Name, "/") {
// 			AuditResourceList["core/v1"] = append(AuditResourceList["core/v1"], apiResource.Name)
// 		}
// 	}

// 	// msg := "ClusterManager is created"
// 	// return clm, msg, http.StatusOK
// }

// func ListAPIResource(apiGroup *metav1.APIGroup) {
// 	reverseMap := make(map[string]string)

// 	// preference first
// 	apiResourceList := &metav1.APIResourceList{}
// 	preferredVersionPath := strings.Replace("/apis/{GROUPVERSION}", "{GROUPVERSION}", apiGroup.PreferredVersion.GroupVersion, -1)
// 	data, err := Clientset.RESTClient().Get().AbsPath(preferredVersionPath).DoRaw(context.TODO())
// 	if err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}
// 	if err := json.Unmarshal(data, apiResourceList); err != nil {
// 		klog.V(1).Infoln(err)
// 		panic(err)
// 	}

// 	for _, apiResource := range apiResourceList.APIResources {
// 		if !strings.Contains(apiResource.Name, "/") {
// 			reverseMap[apiResource.Name] = apiGroup.PreferredVersion.GroupVersion
// 		}
// 	}

// 	// another version
// 	for _, version := range apiGroup.Versions {
// 		if version.GroupVersion == apiGroup.PreferredVersion.GroupVersion {
// 			continue
// 		}
// 		apiResourceList := &metav1.APIResourceList{}
// 		path := strings.Replace("/apis/{GROUPVERSION}", "{GROUPVERSION}", version.GroupVersion, -1)
// 		data, err := Clientset.RESTClient().Get().AbsPath(path).DoRaw(context.TODO())
// 		if err != nil {
// 			klog.V(1).Infoln(err)
// 			panic(err)
// 		}
// 		if err := json.Unmarshal(data, apiResourceList); err != nil {
// 			klog.V(1).Infoln(err)
// 			panic(err)
// 		}

// 		for _, apiResource := range apiResourceList.APIResources {
// 			if !strings.Contains(apiResource.Name, "/") {
// 				if _, ok := reverseMap[apiResource.Name]; !ok {
// 					reverseMap[apiResource.Name] = version.GroupVersion
// 				}
// 			}
// 		}
// 	}

// 	// reverse
// 	for k, v := range reverseMap {
// 		AuditResourceList[v] = append(AuditResourceList[v], k)
// 	}

// }

// func UpdateClusterManager(userId string, userGroups []string, clm *clusterv1alpha1.ClusterManager) (*clusterv1alpha1.ClusterManager, error) {
// 	clmUpdateRuleResult, err := CreateSubjectAccessReview(userId, userGroups, util.CLUSTER_API_GROUP, "clustermanagers", "", clm.Name, "update")
// 	if err != nil {
// 		klog.V(1).Infoln(err)
// 		return nil, err
// 	}

// 	if clmUpdateRuleResult.Status.Allowed {
// 		result, err := customClientset.ClusterV1alpha1().ClusterManagers(clm.Namespace).UpdateStatus(context.TODO(), clm, metav1.UpdateOptions{})
// 		if err != nil {
// 			klog.V(1).Infoln("Update member list in cluster [ " + clm.Name + " ] Failed")
// 			return nil, err
// 		} else {
// 			msg := "Update member list in cluster [ " + clm.Name + " ] Success"
// 			klog.V(3).Infoln(msg)
// 			return result, nil
// 		}
// 	} else {
// 		newErr := errors.NewBadRequest(" User [ " + userId + " ] is not a cluster admin, Cannot invite members")
// 		klog.V(1).Infoln(newErr)
// 		return nil, newErr
// 	}
// }
