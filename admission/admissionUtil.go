package admission

import (
	"strings"

	"k8s.io/api/admission/v1beta1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type patchOps struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func createPatch(po *[]patchOps, o, p string, v interface{}) {
	*po = append(*po, patchOps{
		Op:    o,
		Path:  p,
		Value: v,
	})
}

func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			Result: &metav1.Status{
				Message: "Pass this mutating webhook.",
			},
		}
	}
}

func buildContainerPatch(oldContainerList []corev1.Container, SidecarContainerImage string, logRootPath string) []corev1.Container {
	volumeMounts := []corev1.VolumeMount{}
	// Build volumeMount for sidecar container
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "shared",
		MountPath: "/shared",
		ReadOnly:  true,
	})

	// TO DO
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "fluent-bit-config",
		MountPath: "/fluent-bit/etc/fluent-bit.conf",
		SubPath:   "fluent-bit.conf",
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "fluent-bit-config",
		MountPath: "/fluent-bit/etc/user-parser.conf",
		SubPath:   "user-parser.conf",
	})

	newContainerList := []corev1.Container{}

	for _, container := range oldContainerList {
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      "shared",
			MountPath: logRootPath,
		})
		newContainerList = append(newContainerList, container)
	}

	containerPatch := append(newContainerList, corev1.Container{
		Name:         "fluent-bit",
		Image:        SidecarContainerImage,
		VolumeMounts: volumeMounts,
	})
	return containerPatch
}

func buildSharedVolumePatch() corev1.Volume {

	sharedVolumePatch := corev1.Volume{
		Name: "shared",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	return sharedVolumePatch
}

func buildConfigmapVolumePatch(configMapName string) corev1.Volume {

	// Patch for pod volumes... (configmap volume)
	configmapVolumePatch := corev1.Volume{
		Name: "fluent-bit-config",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
	return configmapVolumePatch
}

func isSystemRequest(userInfo authv1.UserInfo) bool {
	for _, group := range userInfo.Groups {
		gorupElement := strings.Split(group, ":")
		if gorupElement[0] == "system" && ((gorupElement[1] != "masters") && (gorupElement[1] != "authenticated")) { //for kubectl
			return true
		}
	}

	userNameElement := strings.Split(userInfo.Username, ":")
	if userNameElement[0] == "system" {
		return true
	}
	return false
}
