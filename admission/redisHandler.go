package admission

import (
	"encoding/json"
	"strings"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog"
)

func InjectAntiAffinity(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	pod := corev1.Pod{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		return ToAdmissionResponse(err)
	}

	var patch []patchOps

	// 클러스터 고유 이름 파싱
	key := pod.Name
	last := strings.LastIndex(key, "-")
	key = key[:last]
	last_last := strings.LastIndex(key, "-")
	podName := key[:last_last]
	clusterName := podName + "-" + "component"

	// 숫자 파싱
	number := string(pod.Name[len(pod.Name)-1])

	// 1. Label 추가
	lb := pod.Labels
	lb[clusterName] = number
	createPatch(&patch, "add", "/metadata/labels", lb)

	// 2. Leader AntiAffinity 추가
	if lb["role"] == "leader" {
		leader_anti := corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								clusterName: number,
							},
						},
						Namespaces: []string{
							pod.Namespace,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 1,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"role": "leader",
								},
							},
							Namespaces: []string{
								pod.Namespace,
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
		createPatch(&patch, "add", "/spec/affinity", leader_anti)
	}

	// 3. Follower AntiAffinity 추가
	if lb["role"] == "follower" {
		follower_anti := corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								clusterName: number,
							},
						},
						Namespaces: []string{
							pod.Namespace,
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 1,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"role": "follower",
								},
							},
							Namespaces: []string{
								pod.Namespace,
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
		createPatch(&patch, "add", "/spec/affinity", follower_anti)
	}

	if patchData, err := json.Marshal(patch); err != nil {
		return ToAdmissionResponse(err) //msg: error
	} else {
		klog.Infof("JsonPatch=%s", string(patchData))
		reviewResponse.Patch = patchData
	}

	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Allowed = true

	return &reviewResponse
}
