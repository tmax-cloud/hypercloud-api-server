package util

import (
	"encoding/json"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"
)

//Jsonpatch를 담을 수 있는 구조체
type PatchOps struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Jsonpatch를 하나 만들어서 slice에 추가하는 함수
func CreatePatch(po *[]PatchOps, o, p string, v interface{}) {
	*po = append(*po, PatchOps{
		Op:    o,
		Path:  p,
		Value: v,
	})
}

// Response.result.message에 err 메시지 넣고 반환
func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func SetResponse(res http.ResponseWriter, outString string, outJson interface{}, status int) http.ResponseWriter {

	//set Cors
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	res.Header().Set("Access-Control-Max-Age", "3628800");
	res.Header().Set("Access-Control-Expose-Headers", "Content-Type, X-Requested-With, Accept, Authorization, Referer, User-Agent")

	//set StatusCode
	res.WriteHeader(status)

	//set Out
	if outJson!=nil {
		res.Header().Set("Content-Type", "application/json")
		js, err := json.Marshal(outJson)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		res.Write(js)
	} else {
		res.Write([]byte(outString))
	}
	return res
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[item]
	return ok
}

func MonthToInt ( month time.Month ) int {
	switch month {
	case time.January :
		return 1
	case time.February :
		return 2
	case time.March :
		return 3
	case time.April :
		return 4
	case time.May :
		return 5
	case time.June :
		return 6
	case time.July :
		return 7
	case time.August :
		return 8
	case time.September :
		return 9
	case time.October :
		return 10
	case time.November :
		return 11
	case time.December :
		return 12
	default :
		return 0
	}
}
