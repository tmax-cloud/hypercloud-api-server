package grafana

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tmax-cloud/hypercloud-api-server/util"
	"k8s.io/klog"
)

type QueryResponse struct {
	Target     string    `json:"target"`
	Datapoints [][]int64 `json:"datapoints"` // [ ["value", "timestamp in milliseconds"], ["value", ...], ...]
}

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

	var resultJson map[string]interface{}
	bytes, _ := ioutil.ReadAll(req.Body)
	str := string(bytes)
	//klog.Infoln("Result string :", str)
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		klog.Errorln("Error occured during unmarshaling request body")
		util.SetResponse(res, "", nil, http.StatusInternalServerError)
	}
	targets := resultJson["targets"].([]interface{})
	klog.Infoln("targets =", targets)

	for i := range targets {
		//temp := targets[i]
		klog.Infoln(targets[i])
		//targets[i]["target"]
		//klog.Infoln(temp["target"].(string))
	}

	var qr []QueryResponse
	var temp QueryResponse
	temp.Target = "billing_by_account"
	time := time.Now().Unix()
	time *= 1000
	temp.Datapoints = append(temp.Datapoints, []int64{3, (time - 10000*1000)})
	temp.Datapoints = append(temp.Datapoints, []int64{1, (time)})
	qr = append(qr, temp)

	temp = QueryResponse{}
	temp.Target = "billing_by_region"
	temp.Datapoints = append(temp.Datapoints, []int64{8, (time - 1000*1000)})
	temp.Datapoints = append(temp.Datapoints, []int64{3, (time)})
	qr = append(qr, temp)

	klog.Infoln("qr =", qr)

	util.SetResponse(res, "", qr, http.StatusOK)
}

func Annotations(res http.ResponseWriter, req *http.Request) {

}
