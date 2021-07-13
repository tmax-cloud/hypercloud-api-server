package admission

import (
	"encoding/json"
	"errors"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type Meta struct {
	metav1.TypeMeta   `json:",inline"`                                                 // kind & apigroup
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"` // annotation
}

func AddResourceMeta(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	reviewResponse := v1beta1.AdmissionResponse{}

	loc, _ := time.LoadLocation("Local")
	currentTime := time.Now().In(loc)

	userName := ar.Request.UserInfo.Username
	operation := string(ar.Request.Operation)
	ms := Meta{}
	oms := Meta{}
	diff := Meta{}

	// create면.. ownerRef 있는지 확인..

	// First, unmarshal original manifests
	if err := json.Unmarshal(ar.Request.Object.Raw, &ms); err != nil {
		return ToAdmissionResponse(err) //msg: error
	}

	// unmarshal old manifests and diff between ori and old manifests, if exists
	if len(ar.Request.OldObject.Raw) > 0 {
		if err := json.Unmarshal(ar.Request.OldObject.Raw, &oms); err != nil {
			return ToAdmissionResponse(err) //msg: error
		}
		if mergePatch, err := jsonpatch.CreateMergePatch(ar.Request.OldObject.Raw, ar.Request.Object.Raw); err != nil {
			return ToAdmissionResponse(err) //msg: error
		} else {
			if err := json.Unmarshal(mergePatch, &diff); err != nil {
				return ToAdmissionResponse(err) //msg: error
			}
		}
	}

	// Check, whether a request is made by user or system
	// if made by a system, then pass request validation
	// else if req. is made by a user, do request validation
	if !isSystemRequest(userName) {
		if err := denyReq(ms, diff, operation, userName); err != nil {
			return ToAdmissionResponse(err) //msg: error
		}
	}

	var patch []patchOps

	if ms.Annotations == nil {
		am := map[string]interface{}{
			"creator":     userName,
			"createdTime": currentTime,
			"updater":     userName,
			"updatedTime": currentTime,
		}
		createPatch(&patch, "add", "/metadata/annotations", am)
	} else {
		if _, ok := ms.Annotations["creator"]; !ok {
			createPatch(&patch, "add", "/metadata/annotations/creator", userName)
		}
		if _, ok := ms.Annotations["createdTime"]; !ok {
			createPatch(&patch, "add", "/metadata/annotations/createdTime", currentTime)
		}
		createPatch(&patch, "add", "/metadata/annotations/updater", userName)
		createPatch(&patch, "add", "/metadata/annotations/updatedTime", currentTime)
	}

	if patchData, err := json.Marshal(patch); err != nil {
		return ToAdmissionResponse(err) //msg: error
	} else {
		// klog.Infof("JsonPatch=%s", string(patchData))
		reviewResponse.Patch = patchData
	}

	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Allowed = true

	return &reviewResponse
}

func denyReq(ms, diff Meta, op, userName string) error {
	if op == "CREATE" {

		if _, ok := ms.Annotations["creator"]; ok {
			return errors.New("Cannot create resource with creator annotation")
		} else if _, ok := ms.Annotations["createdTime"]; ok {
			return errors.New("Cannot create resource with createdTime annotation")
		} else if _, ok := ms.Annotations["updater"]; ok {
			return errors.New("Cannot create resource with updater annotation")
		} else if _, ok := ms.Annotations["updatedTime"]; ok {
			return errors.New("Cannot create resource with updatedTime annotation")
		}
	}

	if op == "UPDATE" {
		if _, ok := diff.Annotations["creator"]; ok {
			return errors.New("Cannot update resource with creator annotation")
		} else if _, ok := diff.Annotations["createdTime"]; ok {
			return errors.New("Cannot update resource with createdTime annotation")
		}
	}

	return nil
}
