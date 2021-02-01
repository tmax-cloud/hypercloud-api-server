package admission

import (
	"encoding/json"
	"errors"

	k8sApiCaller "github.com/tmax-cloud/hypercloud-api-server/util/Caller"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/klog"
)

var SidecarContainerImage string

func InjectionForPod(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	pod := corev1.Pod{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := pod.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(pod.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := pod.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if pod.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/volumes/-", configmapVolumePatch)

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

func InjectionForDeploy(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	deploy := appsv1.Deployment{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &deploy); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := deploy.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(errors.New("Log collector configuration is empty."))
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(deploy.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := deploy.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if deploy.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForRs(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	rs := appsv1.ReplicaSet{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &rs); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := rs.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(rs.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := rs.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if rs.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForSts(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	sts := appsv1.StatefulSet{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &sts); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := sts.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(sts.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := sts.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if sts.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForDs(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	ds := appsv1.DaemonSet{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &ds); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := ds.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(ds.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := ds.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if ds.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForCj(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	cj := batchv1beta1.CronJob{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &cj); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := cj.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(cj.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := cj.Spec.JobTemplate.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if cj.Spec.JobTemplate.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/jobTemplate/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/jobTemplate/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/jobTemplate/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/jobTemplate/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForJob(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	job := batchv1.Job{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &job); err != nil {
		return ToAdmissionResponse(err)
	}

	var configName string
	if val, exist := job.Labels["tmax.io/log-collector-configuration"]; exist {
		configName = val
	} else {
		err := errors.New("Log collector configuration is empty.")
		klog.Error(err)
		return ToAdmissionResponse(err)
	}
	var logRootPath string
	if res, err := k8sApiCaller.GetFbc(job.GetNamespace(), configName); err != nil {
		klog.Error(err)
		return ToAdmissionResponse(err)
	} else {
		logRootPath = res.Status.LogRootPath
	}

	oldContainerList := job.Spec.Template.Spec.Containers
	containerPatch := buildContainerPatch(oldContainerList, SidecarContainerImage, logRootPath)
	sharedVolumePatch := buildSharedVolumePatch()
	configmapVolumePatch := buildConfigmapVolumePatch(configName)

	var patch []patchOps
	if job.Spec.Template.Spec.Volumes == nil {
		createPatch(&patch, "add", "/spec/template/spec/volumes", []corev1.Volume{})
	}
	createPatch(&patch, "add", "/spec/template/spec/containers", containerPatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", sharedVolumePatch)
	createPatch(&patch, "add", "/spec/template/spec/volumes/-", configmapVolumePatch)

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

func InjectionForTest(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}
	deploy := appsv1.Deployment{}
	if err := json.Unmarshal(ar.Request.Object.Raw, &deploy); err != nil {
		return ToAdmissionResponse(err)
	}

	klog.Info(string(ar.Request.Object.Raw))

	reviewResponse.Allowed = true

	return &reviewResponse
}
