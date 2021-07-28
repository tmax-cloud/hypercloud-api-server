package grafana

import (
	"net/http"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

type QueryResponse struct {
	Target string `json:"target"`
	//Datapoints []DataPoint `json:"datapoints"`
	Datapoints [][]float32 `json:"datapoints"`
}

// type DataPoint struct {
// 	Value         float32 `json:"value"`
// 	Unixtimestamp int64   `json:"unixtimestamp"`
// }

func Get(res http.ResponseWriter, req *http.Request) {
	util.SetResponse(res, "Test Success", nil, http.StatusOK)
}

func Search(res http.ResponseWriter, req *http.Request) {
	metrics := []string{
		"billing_by_account",
		"billing_by_region",
		"billing_by_instance",
		"billing_by_metrics",
	}

	util.SetResponse(res, "", metrics, http.StatusOK)
}

func Query(res http.ResponseWriter, req *http.Request) {

	var qr []QueryResponse

	var temp QueryResponse
	temp.Target = "billing_by_account"
	time := time.Now().Unix()

	temp.Datapoints = append(temp.Datapoints, []float32{0.5, float32(time)})

	qr = append(qr, temp)

	klog.Infoln("qr =", qr)

	util.SetResponse(res, "", qr, http.StatusOK)
}

func Annotations(res http.ResponseWriter, req *http.Request) {

}
